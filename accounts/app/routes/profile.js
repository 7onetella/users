/*eslint no-console: ["error", { allow: ["log", "warn", "error"] }] */
import Route from '@ember/routing/route';
import { inject } from '@ember/service';

export default Route.extend({
  session: inject('session'),

  model() {
    console.log('routes/profile.js: model()')
    console.log("> session.isAuthenticated: " + this.session.isAuthenticated);

    var user_id = this.session.session.content.authenticated.tokenData.user_id
    return this.store.findRecord('user', user_id);
  }
});
