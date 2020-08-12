import Model, { attr } from '@ember-data/model';

export default class UserModel extends Model {
  @attr('string') email;
  @attr('string') password;
  @attr('string') firstname;
  @attr('string') lastname;
  @attr('boolean') mfaenabled;
}
