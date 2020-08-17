import EmberRouter from '@ember/routing/router';
import config from './config/environment';

const Router = EmberRouter.extend({
  location: config.locationType,
  rootURL: config.rootURL
});

Router.map(function() {
  this.route('layout', {path: '/'}, function(){
    this.route('signup', {resetNamespace: true});
    this.route('profile', {resetNamespace: true});
    this.route('security', {resetNamespace: true});
    this.route('totp', {resetNamespace: true});
    this.route('webauthn', {resetNamespace: true});
    this.route('index', {resetNamespace: true, path: '/'});
    this.route('about', {resetNamespace: true});
    this.route('session-expired', {resetNamespace: true});
  });

  this.route('signin');
  this.route('totp-signin');
  this.route('webauthn-signin');
});

export default Router;
