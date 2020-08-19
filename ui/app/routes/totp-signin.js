/*eslint no-console: ["error", { allow: ["warn", "error"] }] */
import Route from '@ember/routing/route';
import { storageFor } from 'ember-local-storage';

export default Route.extend({
  event: storageFor('datastore'),

  model() {
    return {
      'totp': ''
    };
  }
});
