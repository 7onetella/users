import { module, test } from 'qunit';
import { setupTest } from 'ember-qunit';

module('Unit | Route | passwords/new', function(hooks) {
  setupTest(hooks);

  test('it exists', function(assert) {
    let route = this.owner.lookup('route:passwords/new');
    assert.ok(route);
  });
});
