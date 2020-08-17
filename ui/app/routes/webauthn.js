/*eslint no-console: ["error", { allow: ["log", "warn", "error"] }] */
import Route from '@ember/routing/route';
import { inject } from '@ember/service';
import ENV from '../config/environment';

export default Route.extend({
  session: inject('session'),

  model() {
    console.log('routes/webauthn.js: model()')
    if (!this.session.isAuthenticated) {
      console.log('invalid session')
      return {'authenticated': false}
    }

    return {
      'authenticated': true
    };
  }
});
