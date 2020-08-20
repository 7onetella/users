/*eslint no-console: ["error", { allow: ["log", "warn", "error"] }] */
import Route from '@ember/routing/route';
import { inject } from '@ember/service';

export default Route.extend({
  session: inject('session'),

  model() {
    console.log('routes/index.js: model()')

    console.log("  session.isAuthenticated: " + this.session.isAuthenticated);

    if (this.session.isAuthenticated) {
      let user_id = this.session.session.content.authenticated.tokenData.user_id
      let record = this.store.findRecord('user', user_id)
      return {
        'user': record,
        'session': this.session
      };
    }

  }
});
