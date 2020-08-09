import { module, test } from 'qunit';
import { setupTest } from 'ember-qunit';

module('Unit | Route | passwords/delete', function(hooks) {
  setupTest(hooks);

  test('it exists', function(assert) {
    let route = this.owner.lookup('route:passwords/delete');
    assert.ok(route);
  });
});
