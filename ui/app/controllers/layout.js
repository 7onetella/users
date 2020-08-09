/*eslint no-console: ["error", { allow: ["log", "warn", "error"] }] */
import Controller from '@ember/controller';
import { inject } from '@ember/service';
import { get, computed } from '@ember/object';
import config from '../config/environment';

export default Controller.extend({
  router: inject(),
  session: inject('session'),
  userIdle: inject('userIdle'),
  showDebugInfo: false,

  init() {
    this._super(...arguments);
    this.get('userIdle').on('idleChanged', (isIdle) => {
      // isIdle is true if idle. False otherwise.
      console.log('contollers/layout.js: init()')
      console.log("  user idle: " + isIdle)
      if (isIdle === true) {
        this.get('session').invalidate();
        this.get('router').transitionTo('signin');
      }
    })
  },

  config: computed(function() {
    return JSON.stringify(get(config, 'ember-simple-auth-token'), null, '\t');
  }),

  sessionData: computed('session.session.content.authenticated', function() {
    return JSON.stringify(this.get('session.session.content.authenticated'), null, '\t');
  }),

  userData: function () {

  },

  actions: {
    silent_login: function() {
      console.log('controllers/layout.js silent_login()')

      const credentials = { username: 'user8az28y@example.com', password: 'password' }
      const authenticator = 'authenticator:jwt'; // or 'authenticator:jwt'
      let promise = this.get('session').authenticate(authenticator, credentials)

      var that = this
      promise.then(function(){
        console.log("  authentication successful. redirecting to listing page");
        console.log("  router" + that.get('router'))
        that.get('router').transitionTo('security');
      },function(data) {
        // console.log("  data:" + JSON.stringify(data));
        console.log("  reason:" + data.json.reason);
        that.set("loginFailed", true);
        that.set("login_failure_reason", data.json.message)
      });

    }
  }

});
