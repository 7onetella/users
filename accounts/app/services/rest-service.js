/*eslint no-console: ["error", { allow: ["log", "warn", "error"] }] */
import Service from '@ember/service';

export default Service.extend({
  init() {
    this._super(...arguments);
  },

  POST(url, access_token, data, signin_session_token) {
    console.log('> POST url: ' + url)
    let settings = {
      method: 'POST',
      mode: 'cors',
      body: data
    }
    if (signin_session_token) {
      settings.headers = {
        'Content-Type': 'application/json',
        'AuthToken': signin_session_token  // custom header for sending password login auth token
      }
    } else {
      settings.headers = {
        'Content-Type': 'application/json',
        'Authorization': 'Bearer ' + access_token
      }
    }
    return fetch(url, settings);
  }
});
