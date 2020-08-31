/*eslint no-console: ["error", { allow: ["log", "warn", "error"] }] */
import Controller from '@ember/controller';
import { inject } from '@ember/service'
import { storageFor } from 'ember-local-storage';
import ENV from '../config/environment';

export default Controller.extend({
  router: inject(),
  session: inject('session'),
  datastore: storageFor('datastore'),
  web: inject('rest-service'),

  actions: {
    authenticate: function() {
      this.set("login_failed", false);
      console.log('contollers/webauthn-signin.js')

      let BASEURL = ENV.APP.JSONAPIAdaptetHost
      let auth_token = this.get('datastore.auth_token')
      let that = this

      if (!auth_token) {
        that.get('router').transitionTo('login-session-expired');
        return
      }

      this.web.POST(BASEURL+ '/webauthn/login/begin', '', {}, auth_token)
      .then((response) => response.json(), reason => { console.log(reason) })
      .then((credentialRequestOptions) => {
        console.log('> credentialRequestOptions: ' + JSON.stringify(credentialRequestOptions))

        credentialRequestOptions.publicKey.challenge = bufferDecode(credentialRequestOptions.publicKey.challenge);
        credentialRequestOptions.publicKey.allowCredentials.forEach(function (listItem) {
          listItem.id = bufferDecode(listItem.id)
        });

        return navigator.credentials.get({
          publicKey: credentialRequestOptions.publicKey
        })
      })
      .then((assertion) => {
        console.log('> assertion = ' + JSON.stringify(assertion))

        let url = BASEURL + "/webauthn/login/finish"
        let payload = JSON.stringify({
          id: assertion.id,
          rawId: bufferEncode(assertion.rawId),
          type: assertion.type,
          response: {
            authenticatorData: bufferEncode(assertion.response.authenticatorData),
            clientDataJSON: bufferEncode(assertion.response.clientDataJSON),
            signature: bufferEncode(assertion.response.signature),
            userHandle: bufferEncode(assertion.response.userHandle),
          },
        })

        that.web.POST(url, '', payload, auth_token)
          .then((response) => {
              return response.json()
            }, reason => {
              console.log('reason' + reason)
            }
          )
          .then((data) => {
          console.log('  data = ' + JSON.stringify(data))
          const authenticator = 'authenticator:jwt';
          const credentials = {
            auth_token: this.get('datastore.auth_token'),
            sec_auth_token: data.sec_auth_token
          }

          this.session.authenticate(authenticator, credentials)
            .then(() => {
              console.log("> authentication successful. redirecting to index page");
              if (that.get('datastore.client_id')) {
                that.get('router').transitionTo('consent');
              } else {
                that.get('router').transitionTo('index');
              }
            }, data => {
              console.log("> data:" + JSON.stringify(data));
              if (data.json) {
                let reason = data.json.reason
                let message = data.json.message
                console.log("> reason:" + reason)

                if (reason === 'login_auth_expired') {
                  that.get('router').transitionTo('login-session-expired');
                  that.set("login_failed", true);
                  that.set("login_failure_reason", message)
                }

                if (reason === 'invalid_sec_auth_token') {
                  that.set("login_failed", true);
                  that.set("login_failure_reason", message)
                  return
                }
              }
          });

          return data
        })
      }, reason => {
        console.log("> reason:" + reason);
        // alert(reason);
      })
      .catch((err) => {
        console.log("> error:" + JSON.stringify(err));
        if (err.json) {
          let message = err.json.message
          that.set("login_failed", true);
          that.set("login_failure_reason", message)
        } else {
          console.log('error = ' + err)
          that.set("login_failed", true);
          that.set("login_failure_reason", "Error occurred")
        }
      })

    }
  }

});

// Base64 to ArrayBuffer
function bufferDecode(value) {
  return Uint8Array.from(atob(value), c => c.charCodeAt(0));
}

function bufferEncode(value) {
  return btoa(String.fromCharCode.apply(null, new Uint8Array(value)))
    .replace(/\+/g, "-")
    .replace(/\//g, "_")
    .replace(/=/g, "");
}

