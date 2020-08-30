/*eslint no-console: ["error", { allow: ["log", "warn", "error"] }] */
import Controller from '@ember/controller';
import { inject } from '@ember/service'
// import $ from 'jquery';
import ENV from '../config/environment';

export default Controller.extend({
  router: inject(),
  session: inject('session'),
  userService: inject('user-service'),

  actions: {
    confirm: async function(data) {
      console.log('controllers/totp.js confirm()')
      console.log('> totp = ' + data.totp)
      let session_token = this.session.session.content.authenticated.token

      let that = this
      let totpconirmurl = ENV.APP.JSONAPIAdaptetHost + "/totp/confirm"
      let response = await fetch(totpconirmurl, {
        method: 'POST',
        mode: 'cors',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': 'Bearer ' + session_token
        },
        body: JSON.stringify({ totp : data.totp })
      });
      let resp = await response.json();

      if (response.ok) {
        console.log("> fetch()");
        console.log("> status: "+response.status+"\n> data: "+ JSON.stringify(resp));
        that.router.transitionTo('totp-success');
      } else {
        console.log('> error: ' + JSON.stringify(resp));
        that.set("totp_validation_failed", true);
        that.set("totp_validation_message", "Invalid TOTP")
      }

    }
  }
});
