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
        var code = data.json.code
        var message = data.json.message
        var signin_session_token = data.json.signin_session_token
        console.log("> message:" + message)
        // totp auth requested
        if (code === 4800) {
          that.set('datastore.signin_session_token', signin_session_token)
          that.router.transitionTo('totp-signin');
        }
        // webauthn auth required
        if (code === 4900) {
          that.set('datastore.signin_session_token', signin_session_token)
          that.router.transitionTo('webauthn-signin');
        } // invalid password
        if (code === 4300) {
          that.set("login_failed", true);
          that.set("login_failure_reason", message)
        }
      });

    }
  }

});
