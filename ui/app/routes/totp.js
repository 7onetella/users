/*eslint no-console: ["error", { allow: ["log", "warn", "error"] }] */
import Route from '@ember/routing/route';
import { inject } from '@ember/service';

export default Route.extend({
  session: inject('session'),

  model() {
    console.log('routes/totp.js: model()')
    if (!this.get('session.isAuthenticated')) {
      console.log('invalid session')
      return {'authenticated': false}
    }

    var token = this.get('session.session.content.authenticated.token')
    console.log('  token = ' + token)

    return {'token': token, 'totp': '', 'authenticated': true};
  }
});
