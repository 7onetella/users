/*eslint no-console: ["error", { allow: ["log", "warn", "error"] }] */
import Controller from '@ember/controller';
import { inject } from '@ember/service'

export default Controller.extend({
  router: inject(),
  session: inject('session'),
  userService: inject('user-service'),

  actions: {
    toggle_totp: function() {
      console.log('controllers/security.js toggle_totp()')

      var user_id = this.session.session.content.authenticated.tokenData.sub

      var that = this
      this.store.findRecord('user', user_id).then(function(record) {
        var totpenabled = record.get("totpenabled")
        if (totpenabled) {
          record.set("totpenabled", false)
          record.save();
          that.get('router').transitionTo('index');
        } else {
          that.get('router').transitionTo('/totp');
        }
      });

    },
    toggle_webauthn: function() {
      console.log('controllers/security.js toggle_webauthn()')

      var user_id = this.session.session.content.authenticated.tokenData.sub

      var that = this
      this.store.findRecord('user', user_id).then(function(record) {
        var webauthnenabled = record.get("webauthnenabled")
        if (webauthnenabled) {
          record.set("webauthnenabled", false)
          record.save();
          that.get('router').transitionTo('index');
        } else {
          that.get('router').transitionTo('webauthn');
        }
      });
    }
  }
});
