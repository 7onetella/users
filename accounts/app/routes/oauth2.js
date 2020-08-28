/*eslint no-console: ["error", { allow: ["log", "warn", "error"] }] */
import Route from '@ember/routing/route';
import { inject } from '@ember/service';
import {storageFor} from "ember-local-storage";

export default Route.extend({
  session: inject('session'),
  datastore: storageFor('datastore'),
  queryParams: {},

  afterModel(model, transition) {
    console.log('routes/oauth2.js: afterModel()')

    if (this.session.isAuthenticated) {
      console.log('already authenticated')
      this.transitionTo('consent'); // Implicitly aborts the on-going transition.
    }
  },

  model(params) {
    console.log('routes/oauth2.js: model()')
    console.log("> session.isAuthenticated: " + this.session.isAuthenticated);
    console.log("> params = " + JSON.stringify(params))
    this.set('datastore.client_id', params.client_id)
    this.set('datastore.redirect_uri', params.redirect_uri)
    this.set('datastore.scope', params.scope)
    this.set('datastore.response_type', params.response_type)
    this.set('datastore.response_mode', params.response_mode)
    this.set('datastore.nonce', params.nonce)
    this.set('datastore.state', params.state)
    return {}
  },

  resetController(controller, isExiting, transition) {
    // if (isExiting) {
    //   // isExiting would be false if only the route's model was changing
    //   controller.set('client_id', '');
    // }
  }
});