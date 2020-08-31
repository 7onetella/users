/*eslint no-console: ["error", { allow: ["warn", "error"] }] */
import Route from '@ember/routing/route';
import ENV from '../config/environment';

export default Route.extend({
  model() {
    return {
      'apihost': ENV.APP.JSONAPIAdaptetHost
    };
  }
});
