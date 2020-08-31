/*eslint no-console: ["error", { allow: ["log", "warn", "error"] }] */
import Controller from '@ember/controller';
import { inject } from '@ember/service'
import ENV from '../config/environment';

export default Controller.extend({
  router: inject(),
  session: inject('session'),
  userService: inject('user-service'),
  web: inject('rest-service'),

  actions: {
    register: async function() {
      console.log('controllers/webauthn.js register()')
      let BASEURL = ENV.APP.JSONAPIAdaptetHost
      let session_token = this.session.session.content.authenticated.token
      let that = this

      this.web.POST(BASEURL + "/webauthn/register/begin", session_token, {})
        .then(response => response.json())
        .then((credentialCreationOptions) => {
          console.log("> credentialCreationOptions = " + JSON.stringify(credentialCreationOptions))
          credentialCreationOptions.publicKey.challenge = bufferDecode(credentialCreationOptions.publicKey.challenge);
          credentialCreationOptions.publicKey.user.id = bufferDecode(credentialCreationOptions.publicKey.user.id);

          return navigator.credentials.create({
            publicKey: credentialCreationOptions.publicKey
          });
        }).then((credential) => {
          console.log("> then(credential)")
          let attestationObject = credential.response.attestationObject;
          let clientDataJSON = credential.response.clientDataJSON;
          let rawId = credential.rawId;

          return this.web.POST(BASEURL + "/webauthn/register/finish", session_token, JSON.stringify({
            id: credential.id,
            rawId: bufferEncode(rawId),
            type: credential.type,
            response: {
              attestationObject: bufferEncode(attestationObject),
              clientDataJSON: bufferEncode(clientDataJSON),
            },
          }));
        }, reason => {
          console.log('reason = ' + reason)
          alert('reason =>' + reason)
          return { ok: false }
        }).then((response) => {
          if (response && response.ok) {
            console.log('response = ' + response.ok)
            that.router.transitionTo('webauthn-success');
          }
        }).catch((err) => {
          alert(err)
          return
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
