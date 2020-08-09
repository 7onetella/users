/*eslint no-console: ["error", { allow: ["warn", "error"] }] */
import Component from '@ember/component';
import {inject as service} from '@ember/service'
import { get } from '@ember/object';

export default Component.extend({
  router: service(),
  // passwords: service(),
  store: service(),
  session: service(),
  flashMessages: service(),
  searchtext: "",

  actions: {
    edit(password) {
    },
    onSuccess() {
      get(this, 'flashMessages').success('copied password to clipboard');
    }
  }
});
