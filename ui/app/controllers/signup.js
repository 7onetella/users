/*eslint no-console: ["error", { allow: ["log", "warn", "error"] }] */
import Controller from '@ember/controller';
import { inject } from '@ember/service'

export default Controller.extend({
  router: inject(),

  actions: {
    signup(user) {
      console.log("controllers/signup.js")
      console.log("  user = " + user.firstname)
      let record = this.store.createRecord('user', user);
      let promise = record.save(); // => POST to '/users'

      var that = this;
      promise.then(function() {
        that.get('router').transitionTo('/signin')
      },function(err) {
        console.warn("  reason:" + err);
        that.set("singup_failed", true);
      });

    }
  }
});
