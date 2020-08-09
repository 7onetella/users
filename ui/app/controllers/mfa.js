/*eslint no-console: ["error", { allow: ["log", "warn", "error"] }] */
import Controller from '@ember/controller';
import { inject } from '@ember/service'

export default Controller.extend({
  router: inject(),
  session: inject('session'),
  userService: inject('user-service'),

  actions: {
    confirm: function(data) {
      console.log('controllers/mfa.js confirm()')

      console.log('  otop = ' + data.totp)
      var session_token = this.get('session.session.content.authenticated.token')

      $.ajax({
        url: "http://localhost:8080/mfa/confirm",
        type: 'post',
        dataType: 'json',
        data: JSON.stringify({ totp : data.totp }),
        async: true,
        crossDomain: 'true',
        beforeSend: function(xhr){xhr.setRequestHeader('Authorization', 'Bearer ' + session_token);},
        success: function(data, status) {
          console.log("Status: "+status+"\nData: "+data);
        },
        error: function(error, txtStatus) {
          console.log(txtStatus);
          console.log('error');
        }})

      this.get('router').transitionTo('index');
    }
  }
});
