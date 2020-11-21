/*eslint no-console: ["error", { allow: ["log", "warn", "error"] }] */
import Controller from '@ember/controller';
import { inject } from '@ember/service'
import { storageFor } from 'ember-local-storage';

export default Controller.extend({
  router: inject(),
  session: inject('session'),
  datastore: storageFor('datastore'),

  actions: {
    authenticate: function(data) {
      console.log('contollers/totp.js')
      this.set("login_failed", false);

      const authenticator = 'authenticator:jwt';
      const credentials = {
        signin_session_token: this.get('datastore.signin_session_token'),
        totp: data.totp
      }
      let promise = this.session.authenticate(authenticator, credentials)

      var that = this
      promise.then(function(){
        console.log("> authentication successful. redirecting to index page");
        if (that.get('datastore.client_id')) {
          that.get('router').transitionTo('consent');
        } else {
          that.get('router').transitionTo('index');
        }
      },function(data) {
        console.log("> data:" + JSON.stringify(data));
        if (data.status && data.status == 500) {
          that.set("login_failed", true);
          that.set("login_failure_reason", data.statusText)
        }
        if (data.json) {
          var code = data.json.code
          var message = data.json.message
          console.log("> message:" + message)
          // totp code is invalid
          if (code === 4500) {
            that.set("login_failed", true);
            that.set("login_failure_reason", message)
          }
          // signin session expired
          if (code === 4200) {
            that.get('router').transitionTo('login-session-expired');
          }
        }
      });
    }
  }

});
