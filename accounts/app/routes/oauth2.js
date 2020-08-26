/*eslint no-console: ["error", { allow: ["log", "warn", "error"] }] */
import Route from '@ember/routing/route';
import { inject } from '@ember/service';

export default Route.extend({
  session: inject('session'),
  queryParams: {},

  model(params) {
    console.log('routes/oauth2.js: model()')
    console.log("> session.isAuthenticated: " + this.session.isAuthenticated);

    console.log("> model = " + JSON.stringify(params))

    if (!this.session.isAuthenticated) {
      console.log('invalid session')
      return {'session': this.session}
    }

    return {}
  },

  resetController(controller, isExiting, transition) {
    if (isExiting) {
      // isExiting would be false if only the route's model was changing
      controller.set('client_id', '');
    }
  }
});
