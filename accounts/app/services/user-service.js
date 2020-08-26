import Service, { inject } from '@ember/service';

export default Service.extend({
  store: inject('store'),

  init() {
    this._super(...arguments);
    // console.log(this.get('store'));
  },

  getUser(user_id) {
    return this.store.findRecord('user', user_id);
  }
});
