/*eslint no-console: ["error", { allow: ["log", "warn", "error"] }] */
import Controller from '@ember/controller';
import { inject } from '@ember/service'
import { storageFor } from 'ember-local-storage';

export default Controller.extend({
  router: inject(),
  session: inject('session'),
  datastore: storageFor('datastore'),

  actions: {
    authenticate: function() {
      console.log('contollers/signin.js')
      const credentials = {
        username: this.username,
        password: this.password,
        totp: this.totp
      }
      const authenticator = 'authenticator:jwt';
      let promise = this.session.authenticate(authenticator, credentials)

      var that = this
      promise.then(function(){
        console.log("> authentication successful. redirecting to index page");
        that.router.transitionTo('index');
      }, data => {
        console.log("> data:" + JSON.stringify(data));
        var reason = data.json.reason
        var message = data.json.message
        var auth_token = data.json.auth_token
        console.log("> reason:" + reason)
        if (reason === 'missing_totp') {
          that.set('datastore.auth_token', auth_token)
          that.router.transitionTo('totp-signin');
        }
        if (reason === 'webauthn_required') {
          that.set('datastore.auth_token', auth_token)
          that.router.transitionTo('webauthn-signin');
        }
        if (reason === 'invalid_password') {
          that.set("loginFailed", true);
          that.set("login_failure_reason", message)
        }
      });

    }
  }
});
