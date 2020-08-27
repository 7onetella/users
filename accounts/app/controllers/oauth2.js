/*eslint no-console: ["error", { allow: ["log", "warn", "error"] }] */
import Controller from '@ember/controller';
import { inject } from '@ember/service'

export default Controller.extend({
  queryParams: ["client_id", "redirect_uri", "scope", "response_type", "response_mode", "nonce", "state"],
  actions: {}
});
