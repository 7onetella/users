/*eslint no-console: ["error", { allow: ["log", "warn", "error"] }] */
import Controller from '@ember/controller';
import { inject } from '@ember/service'
import { storageFor } from 'ember-local-storage';

export default Controller.extend({
  router: inject(),
  session: inject('session'),
  event: storageFor('event'),

  actions: {
    authenticate: function(data) {
      console.log('contollers/totp.js')
      const authenticator = 'authenticator:jwt'; // or 'authenticator:jwt'
      const credentials = {
        totp: data.totp,
        event_id: this.get('event.id')
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
    }
  }

});
