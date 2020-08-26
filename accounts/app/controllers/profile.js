/*eslint no-console: ["error", { allow: ["log", "warn", "error"] }] */
import Controller from '@ember/controller';
import { inject } from '@ember/service'

export default Controller.extend({
  router: inject(),
  session: inject('session'),
  userService: inject('user-service'),

  actions: {
    load: function() {
      console.log('controllers/profile.js')

    },
    update: function(user) {
      console.log('controllers/profile.js update()')
      console.log("  user.id = " + user.id)

      var that = this
      this.store.findRecord('user', user.id).then(function(record) {
        record.set("firstname", user.firstname)
        record.set("lastname", user.lastname)
        record.save();
        that.get('router').transitionTo('index');
      });
    }
  }
});
