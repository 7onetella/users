import UserIdleService from 'ember-user-activity/services/user-idle';

export default UserIdleService.extend({
  IDLE_TIMEOUT: 300000 // 30 seconds 
});