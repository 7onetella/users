/*eslint no-console: ["error", { allow: ["log", "warn", "error"] }] */
import Controller from '@ember/controller';
import { inject } from '@ember/service'
import { storageFor } from 'ember-local-storage';
import ENV from '../config/environment';

export default Controller.extend({
  router: inject(),
  session: inject('session'),
  datastore: storageFor('datastore'),

  actions: {
    authenticate: function(data) {
      console.log('contollers/webauthn-signin.js')

      var session_token = this.session.session.content.authenticated.token
      var webauthnpurl = ENV.APP.JSONAPIAdaptetHost + "/webauthn/login/begin"

      var that = this
      var settings = {
        url: webauthnpurl,
        type: 'post',
        dataType: 'json',
        async: true,
        crossDomain: 'true',
        beforeSend: function (xhr) {
          xhr.setRequestHeader('Authorization', 'Bearer ' + session_token);
          xhr.setRequestHeader('AuthToken', that.get('datastore.auth_token'));
        }
      }

      $.ajax(settings).then((credentialRequestOptions) => {
        credentialRequestOptions.publicKey.challenge = bufferDecode(credentialRequestOptions.publicKey.challenge);
        credentialRequestOptions.publicKey.allowCredentials.forEach(function (listItem) {
          listItem.id = bufferDecode(listItem.id)
        });

        return navigator.credentials.get({
          publicKey: credentialRequestOptions.publicKey
        })
      })
      .then((assertion) => {

        let authData = assertion.response.authenticatorData;
        let clientDataJSON = assertion.response.clientDataJSON;
        let rawId = assertion.rawId;
        let sig = assertion.response.signature;
        let userHandle = assertion.response.userHandle;

        settings.url = ENV.APP.JSONAPIAdaptetHost + "/webauthn/login/finish"
        settings.data = JSON.stringify({
          id: assertion.id,
          rawId: bufferEncode(rawId),
          type: assertion.type,
          response: {
            authenticatorData: bufferEncode(authData),
            clientDataJSON: bufferEncode(clientDataJSON),
            signature: bufferEncode(sig),
            userHandle: bufferEncode(userHandle),
          },
        })

        // console.log("  then(assertion) ajax settings = " + JSON.stringify(settings))
        $.ajax(settings).then((data) => {
          console.log('  data = ' + JSON.stringify(data))
          alert("successfully logged in !")
          const authenticator = 'authenticator:jwt';
          const credentials = {
            auth_token: this.get('datastore.auth_token'),
            sec_auth_token: data.sec_auth_token
          }
          let promise = this.session.authenticate(authenticator, credentials)

          var that = this
          promise.then(function(){
            console.log("  authentication successful. redirecting to index page");
            console.log("  router" + that.get('router'))
            that.get('router').transitionTo('index');
          },function(data) {
            console.log("  data:" + JSON.stringify(data));
            var reason = data.json.reason
            var message = data.json.message
            console.log("  reason:" + reason)
            if (reason === 'invalid_totp') {
              that.set("loginFailed", true);
              that.set("login_failure_reason", message)
            }
          });

          return data
        })
      })
      .catch((error) => {
        console.log(error)
        alert("failed to login")
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

