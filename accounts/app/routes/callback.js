/*eslint no-console: ["error", { allow: ["log", "warn", "error"] }] */
import Route from '@ember/routing/route';

export default Route.extend({
  queryParams: {},

  beforeModel() {
    console.log('routes/callback.js')
  },

  model() {
    console.log('routes/callback.js')
    return {}
  },
});
