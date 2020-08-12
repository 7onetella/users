/*eslint no-console: ["error", { allow: ["log", "warn", "error"] }] */
import Route from '@ember/routing/route';
import { inject } from '@ember/service';

export default Route.extend({
  router: inject(),
  session: inject('session'),

  setupController (controller, model) {
    console.log('routes/application.js')
    // console.log("session.isAuthenticated: " + this.get('session.isAuthenticated'));
    this._super(controller, model);
  },

  actions: {
    invalidateSession: function() {
      console.log('routes/applications.js: invalidateSession()')
      this.session.invalidate();
      this.router.transitionTo('index');
    }
  }
});
