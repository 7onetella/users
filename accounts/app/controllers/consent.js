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
    authorize: function(data) {
      console.log('contollers/consent.js authorize()')

      let session_token = this.session.session.content.authenticated.token
      let oauthurl = ENV.APP.JSONAPIAdaptetHost + "/oauth2/authorize"

      let settings = {
        url: oauthurl,
        type: 'post',
        dataType: 'json',
        async: true,
        crossDomain: 'true',
        beforeSend: function (xhr) {
          xhr.setRequestHeader('Authorization', 'Bearer ' + session_token);
        }
      }
      settings.data = JSON.stringify({
        client_id: this.get('datastore.client_id'),
        redirect_uri: this.get('datastore.redirect_uri'),
        scope: this.get('datastore.scope'),
        response_type: this.get('datastore.response_type'),
        response_mode: this.get('datastore.response_mode'),
        nonce: this.get('datastore.nonce'),
        state: this.get('datastore.state')
      })

      $.ajax(settings).then((result) => {
        console.log("result = " + JSON.stringify(result))
        window.location.replace(result.redirect_uri + "?code=" + result.code + "&nonce=" + result.nonce + "&state=" + result.state)
      })
    }
  }

});
