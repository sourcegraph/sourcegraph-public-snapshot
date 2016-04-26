describe('Snippet', function() {

  it('amplitude object should exist', function() {
    assert.isObject(window);
    assert.isObject(window.amplitude);
    assert.isFunction(window.amplitude.init);
    assert.isFunction(window.amplitude.logEvent);
  });

  it('amplitude object should proxy functions', function() {
    amplitude.init('API_KEY');
    amplitude.logEvent('Event', {prop: 1});
    assert.lengthOf(amplitude._q, 2);
    assert.deepEqual(amplitude._q[0], ['init', 'API_KEY']);
  });

  it('amplitude object should proxy Identify object and calls', function() {
    var identify = new amplitude.Identify().set('key1', 'value1').unset('key2');
    identify.add('key3', 2).setOnce('key4', 'value2');

    assert.lengthOf(identify._q, 4);
    assert.deepEqual(identify._q[0], ['set', 'key1', 'value1']);
    assert.deepEqual(identify._q[1], ['unset', 'key2']);
    assert.deepEqual(identify._q[2], ['add', 'key3', 2]);
    assert.deepEqual(identify._q[3], ['setOnce', 'key4', 'value2']);
  });

  it('amplitude object should proxy Revenue object and calls', function() {
    var revenue = new amplitude.Revenue().setProductId('productIdentifier').setQuantity(5).setPrice(10.99);
    assert.lengthOf(revenue._q, 3);
    assert.deepEqual(revenue._q[0], ['setProductId', 'productIdentifier']);
    assert.deepEqual(revenue._q[1], ['setQuantity', 5]);
    assert.deepEqual(revenue._q[2], ['setPrice', 10.99]);
  });
});
