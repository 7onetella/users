/*eslint no-console: ["error", { allow: ["log", "warn", "error"] }] */
import Controller from '@ember/controller';
import { inject } from '@ember/service'

export default Controller.extend({
  router: inject(),
  session: inject('session'),
  userService: inject('user-service'),
  queryParams: ["client_id", "redirect_uri", "scope", "response_type", "response_mode", "nonce", "state"],

  actions: {
    load: function() {
      console.log('controllers/oauth2.js')
    },
    update: function(user) {
      console.log('controllers/oauth2.js update()')
      console.log(">  user.id = " + user.id)
    }
  }
});
