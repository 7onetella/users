/*eslint no-console: ["error", { allow: ["log", "warn", "error"] }] */
import Controller from '@ember/controller';
import { inject } from '@ember/service'

export default Controller.extend({
  router: inject(),
  session: inject('session'),
  userService: inject('user-service'),

  actions: {
    toggle_mfa: function() {
      console.log('controllers/security.js toggle_mfa()')

      var user_id = this.get('session.session.content.authenticated.tokenData.user_id')

      var that = this
      this.store.findRecord('user', user_id).then(function(record) {
        var mfaenabled = record.get("mfaenabled")
        if (mfaenabled) {
          record.set("mfaenabled", false)
          record.save();
          that.get('router').transitionTo('index');
        } else {
          that.get('router').transitionTo('/mfa');
        }
      });

    }
  }
});
