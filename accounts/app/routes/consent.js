/*eslint no-console: ["error", { allow: ["log", "warn", "error"] }] */
import Route from '@ember/routing/route';
import { inject } from '@ember/service';
import { storageFor } from "ember-local-storage";

export default Route.extend({
  session: inject('session'),
  datastore: storageFor('datastore'),

  model() {
    console.log('routes/consent.js: model()')
    let s = this.get('datastore.scope')
    console.log("s = " + JSON.stringify(s))
    let scopes = this.get('datastore.scope').split(",")
    return {
      "scopes": scopes
    }
  },

});
