/*eslint no-console: ["error", { allow: ["log", "warn", "error"] }] */
import Route from '@ember/routing/route';

export default Route.extend({
  model() {
    console.log('routes/index.js: model()')
  }
});
