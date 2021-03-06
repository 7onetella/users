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
      let signin_session_token = this.get('datastore.signin_session_token')
      let that = this

      if (!signin_session_token) {
        that.get('router').transitionTo('login-session-expired');
        return
      }

      this.web.POST(BASEURL+ '/webauthn/login/begin', '', {}, signin_session_token)
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

        that.web.POST(url, '', payload, signin_session_token)
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
            signin_session_token: this.get('datastore.signin_session_token'),
            webauthn_session_token: data.webauthn_session_token
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
                let message = data.json.message
                let code = data.json.code
                console.log("> message:" + message)
                // signin session expired
                if (code === 4200) {
                  that.get('router').transitionTo('login-session-expired');
                  that.set("login_failed", true);
                  that.set("login_failure_reason", message)
                }
                // webauthn auth failed
                if (code === 4400) {
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

