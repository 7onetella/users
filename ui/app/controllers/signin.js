/*eslint no-console: ["error", { allow: ["log", "warn", "error"] }] */
import Controller from '@ember/controller';
import { inject } from '@ember/service'

export default Controller.extend({
  router: inject(),
  session: inject('session'),

  actions: {
    authenticate: function() {
      console.log('contollers/signin.js')
      const credentials = this.getProperties('username', 'password', 'totp');
      const authenticator = 'authenticator:jwt'; // or 'authenticator:jwt'
      let promise = this.get('session').authenticate(authenticator, credentials)

      var that = this
      promise.then(function(){
        console.log("  authentication successful. redirecting to index page");
        console.log("  router" + that.get('router'))
        that.get('router').transitionTo('index');
      },function(data) {
        // console.log("  data:" + JSON.stringify(data));
        var reason = data.json.reason
        var message = data.json.message
        console.log("  reason:" + reason)
        if (reason === 'missing_totp') {
          that.get('router').transitionTo('totp');
        }
        if (reason === 'invalid_credential') {
          that.set("loginFailed", true);
          that.set("login_failure_reason", message)
        }
      });

    }
  }
});
