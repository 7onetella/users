/*eslint no-console: ["error", { allow: ["log", "warn", "error"] }] */
import Service from '@ember/service';

export default Service.extend({
  init() {
    this._super(...arguments);
  },

  POST(url, session_token, data) {
    console.log('> POST url: ' + url)
    return fetch(url, {
      method: 'POST',
      mode: 'cors',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': 'Bearer ' + session_token
      },
      body: data
    });
  }
});
