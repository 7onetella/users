/*eslint no-console: ["error", { allow: ["log", "warn", "error"] }] */
import Component from '@ember/component';
import {inject} from '@ember/service'
import {storageFor} from "ember-local-storage";

export default Component.extend({
  router: inject(),
  session: inject('session'),
  datastore: storageFor('datastore'),

  actions: {
    authenticate: function() {
      console.log('components/signin.js')
      // initialize vars
      this.set("login_failed", false);
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
        // clean up variables after successful login
        that.username = ''
        that.password = ''
        if (that.get('datastore.client_id')) {
          that.router.transitionTo('consent');
        } else {
          that.router.transitionTo('index');
        }
      }, data => {
        console.log("> data:" + JSON.stringify(data));
        var reason = data.json.reason
        var message = data.json.message
        var auth_token = data.json.auth_token
        console.log("> reason:" + reason)
        if (reason === 'login_totp_requested') {
          that.set('datastore.auth_token', auth_token)
          that.router.transitionTo('totp-signin');
        }
        if (reason === 'login_webauthn_requested') {
          that.set('datastore.auth_token', auth_token)
          that.router.transitionTo('webauthn-signin');
        }
        if (reason === 'login_password_invalid') {
          that.set("login_failed", true);
          that.set("login_failure_reason", message)
        }
      });

    }
  }

});
