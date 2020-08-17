/*eslint no-console: ["error", { allow: ["log", "warn", "error"] }] */
import Controller from '@ember/controller';
import { inject } from '@ember/service'
import { storageFor } from 'ember-local-storage';

export default Controller.extend({
  router: inject(),
  session: inject('session'),
  event: storageFor('event'),

  actions: {
    authenticate: function() {
      console.log('contollers/signin.js')
      const credentials = {
        username: this.username,
        password: this.password,
        totp: this.totp
      }
      const authenticator = 'authenticator:jwt'; // or 'authenticator:jwt'
      let promise = this.session.authenticate(authenticator, credentials)

      var that = this
      promise.then(function(){
        console.log("  authentication successful. redirecting to index page");
        console.log("  router" + that.get('router'))
        that.router.transitionTo('index');
      },function(data) {
        console.log("  data:" + JSON.stringify(data));
        var reason = data.json.reason
        var message = data.json.message
        var event_id = data.json.event_id
        console.log("  reason:" + reason)
        if (reason === 'missing_totp') {
          that.set('event.id', event_id)
          that.router.transitionTo('totp-signin');
        }
        if (reason === 'webauthn_required') {
          that.set('event.id', event_id)
          that.router.transitionTo('webauthn-signin');
        }
        if (reason === 'invalid_credential') {
          that.set("loginFailed", true);
          that.set("login_failure_reason", message)
        }
      });

    }
  }
});
