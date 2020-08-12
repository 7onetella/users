/*eslint no-console: ["error", { allow: ["log", "warn", "error"] }] */
import Route from '@ember/routing/route';
import { inject } from '@ember/service';

export default Route.extend({
  session: inject('session'),
  userService: inject('user-service'),

  model() {
    if (!this.session.isAuthenticated) {
      console.log('invalid session')
      return {'session': this.session}
    }

    console.log('routes/security.js: model()')
    var token = this.session.session.content.authenticated.token
    console.log('  token = ' + token)

    var user_id = this.session.session.content.authenticated.tokenData.user_id
    console.log('  user_id = ' + user_id)
    // var userStr = JSON.stringify(this.userService.getUser(user_id))
    var record = this.store.findRecord('user', user_id)

    return {
      'token': token,
      'user': record,
      'session': this.session,
    };
  }
});
