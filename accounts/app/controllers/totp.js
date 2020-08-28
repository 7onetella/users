/*eslint no-console: ["error", { allow: ["log", "warn", "error"] }] */
import Controller from '@ember/controller';
import { inject } from '@ember/service'
import $ from 'jquery';
import ENV from '../config/environment';

export default Controller.extend({
  router: inject(),
  session: inject('session'),
  userService: inject('user-service'),

  actions: {
    confirm: function(data) {
      console.log('controllers/totp.js confirm()')
      console.log('> totp = ' + data.totp)
      var session_token = this.session.session.content.authenticated.token

      var that = this
      $.ajax({
        url: ENV.APP.JSONAPIAdaptetHost + "/totp/confirm",
        type: 'post',
        dataType: 'json',
        data: JSON.stringify({ totp : data.totp }),
        async: true,
        crossDomain: 'true',
        beforeSend: function(xhr){xhr.setRequestHeader('Authorization', 'Bearer ' + session_token);},
        success: function(data, status) {
          console.log("> status: "+status+"\n> data: "+data);
          that.router.transitionTo('totp-success');
        },
        error: function(error) {
          console.log('> error: ' + JSON.stringify(error));
          that.set("totp_validation_failed", true);
          that.set("totp_validation_message", "Invalid TOTP")
        }})
    }
  }
});