/*eslint no-console: ["error", { allow: ["log", "warn", "error"] }] */
import Controller from '@ember/controller';
import { inject } from '@ember/service'
import { storageFor } from 'ember-local-storage';
import ENV from '../config/environment';

export default Controller.extend({
  router: inject(),
  session: inject('session'),
  datastore: storageFor('datastore'),

  actions: {
    authorize: async function() {
      console.log('contollers/consent.js authorize()')

      let session_token = this.session.session.content.authenticated.token
      let oauthurl = ENV.APP.JSONAPIAdaptetHost + "/oauth2/authorize"

      let response = await fetch(oauthurl, {
        method: 'POST',
        mode: 'cors',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': 'Bearer ' + session_token
        },
        body: JSON.stringify({
          client_id: this.get('datastore.client_id'),
          redirect_uri: this.get('datastore.redirect_uri'),
          scope: this.get('datastore.scope'),
          response_type: this.get('datastore.response_type'),
          response_mode: this.get('datastore.response_mode'),
          nonce: this.get('datastore.nonce'),
          state: this.get('datastore.state')
        })
      });
      let result = await response.json();

      if (response.ok) {
        console.log("result = " + JSON.stringify(result))
        window.location.replace(result.redirect_uri + "?code=" + result.code + "&nonce=" + result.nonce + "&state=" + result.state)
      }
    }
  }

});
