import TokenAdapterMixin from 'ember-simple-auth-token/mixins/token-adapter';
import ENV from '../config/environment';
import JSONAPIAdapter from '@ember-data/adapter/json-api';

export default JSONAPIAdapter.extend(TokenAdapterMixin, {
  host: ENV.APP.JSONAPIAdaptetHost,
  namespace: ''
});
