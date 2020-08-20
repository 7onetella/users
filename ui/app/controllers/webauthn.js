/*eslint no-console: ["error", { allow: ["log", "warn", "error"] }] */
import Controller from '@ember/controller';
import { inject } from '@ember/service'
import $ from 'jquery';
import ENV from '../config/environment';

export default Controller.extend({
  router: inject(),
  session: inject('session'),
  userService: inject('user-service'),

  actions: {
    register: function(data) {
      console.log('controllers/webauthn.js register()')

      var session_token = this.session.session.content.authenticated.token
      var webauthnpurl = ENV.APP.JSONAPIAdaptetHost + "/webauthn/register/begin"

      var settings = {
        url: webauthnpurl,
        type: 'post',
        dataType: 'json',
        async: true,
        crossDomain: 'true',
        beforeSend: function (xhr) {
          xhr.setRequestHeader('Authorization', 'Bearer ' + session_token);
        }
      }

      $.ajax(settings).then((credentialCreationOptions) => {
        console.log("  credentialCreationOptions = " + JSON.stringify(credentialCreationOptions))

        credentialCreationOptions.publicKey.challenge = bufferDecode(credentialCreationOptions.publicKey.challenge);
        credentialCreationOptions.publicKey.user.id = bufferDecode(credentialCreationOptions.publicKey.user.id);

        return navigator.credentials.create({
          publicKey: credentialCreationOptions.publicKey
        })
      })
      .then((credential) => {
        console.log("  then(credential)")
        let attestationObject = credential.response.attestationObject;
        let clientDataJSON = credential.response.clientDataJSON;
        let rawId = credential.rawId;

        settings.url = ENV.APP.JSONAPIAdaptetHost + "/webauthn/register/finish"
        settings.data = JSON.stringify({
          id: credential.id,
          rawId: bufferEncode(rawId),
          type: credential.type,
          response: {
            attestationObject: bufferEncode(attestationObject),
            clientDataJSON: bufferEncode(clientDataJSON),
          },
        })
        console.log("  then(credential) ajax settings = " + JSON.stringify(settings))
        $.ajax(settings)
      })
      .then((success) => {
        this.router.transitionTo('webauthn-success');
        return
      })
      .catch((error) => {
        console.log(error)
        alert("failed to register ")
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
