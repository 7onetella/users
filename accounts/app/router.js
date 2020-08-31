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
    this.route('webauthn-success', {resetNamespace: true});
    this.route('totp-success', {resetNamespace: true});
  });

  this.route('signin');
  this.route('totp-signin');
  this.route('webauthn-signin');
  this.route('login-session-expired');
  this.route('oauth2');
  this.route('consent');

  this.route('demo', {path: '/demo'}, function(){
    this.route('oauth2-signin', {resetNamespace: true});
    this.route('oauth2-callback', {resetNamespace: true});
    this.route('oauth2-post-signin', {resetNamespace: true});
  });

});

export default Router;
