import DS from 'ember-data';
import TokenAdapterMixin from 'ember-simple-auth-token/mixins/token-adapter';
import ENV from '../config/environment';

export default DS.JSONAPIAdapter.extend(TokenAdapterMixin, {
  host: ENV.APP.JSONAPIAdaptetHost,
  namespace: ''
});
