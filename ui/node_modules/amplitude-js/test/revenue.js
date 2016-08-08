var Revenue = require('../src/revenue.js');

describe('Revenue', function() {

  it('should initialize fields', function() {
    var revenue = new Revenue();
    assert.equal(revenue._productId, null);
    assert.equal(revenue._quantity, 1);
    assert.equal(revenue._price, null);
    assert.equal(revenue._revenueType, null);
    assert.equal(revenue._properties, null);
  });

  it('should set productId', function() {
    var revenue = new Revenue();
    assert.equal(revenue._productId, null);

    var productId = 'testProductId';
    revenue.setProductId(productId);
    assert.equal(revenue._productId, productId);

    // verify that null and empty strings are ignored
    revenue.setProductId(null);
    assert.equal(revenue._productId, productId);
    revenue.setProductId('');
    assert.equal(revenue._productId, productId);
  });

  it('should set quantity', function() {
    var revenue = new Revenue();
    assert.equal(revenue._quantity, 1);

    var quantity = 15;
    revenue.setQuantity(quantity);
    assert.equal(revenue._quantity, quantity);

    // verify that doubles are rounded down
    revenue.setQuantity(15.75);
    assert.equal(revenue._quantity, quantity);

    // verify that non number values are ignored
    revenue.setQuantity('test');
    assert.equal(revenue._quantity, quantity);
  });

  it('should set price', function() {
    var revenue = new Revenue();
    assert.equal(revenue._price, null);

    var price = 10.99;
    revenue.setPrice(price);
    assert.equal(revenue._price, price);

    // verify that non numbers are ignored
    revenue.setPrice('test');
    assert.equal(revenue._price, price);
  });

  it('should set revenue type', function() {
    var revenue = new Revenue();
    assert.equal(revenue._revenueType, null);

    var revenueType = 'testRevenueType'
    revenue.setRevenueType(revenueType);
    assert.equal(revenue._revenueType, revenueType);

    // verify that non strings and nulls are ignored
    revenue.setRevenueType(15);
    assert.equal(revenue._revenueType, revenueType);
    revenue.setRevenueType(null);
    assert.equal(revenue._revenueType, revenueType);
  });

  it('should set event properties', function() {
    var revenue = new Revenue();
    assert.equal(revenue._properties, null);

    var properties = {'city':'Boston'};
    revenue.setEventProperties(properties);
    assert.deepEqual(revenue._properties, properties);

    // verify that non objects are ignored
    revenue.setEventProperties(null);
    assert.deepEqual(revenue._properties, properties);
    revenue.setEventProperties('test');
    assert.deepEqual(revenue._properties, properties);
  });

  it('should validate revenue objects', function () {
    var revenue = new Revenue();
    assert.isFalse(revenue._isValidRevenue());
    revenue.setProductId('testProductId');
    assert.isFalse(revenue._isValidRevenue());
    revenue.setPrice(10.99);
    assert.isTrue(revenue._isValidRevenue());

    var revenue2 = new Revenue();
    assert.isFalse(revenue2._isValidRevenue());
    revenue2.setPrice(10.99);
    revenue2.setQuantity(15);
    assert.isFalse(revenue2._isValidRevenue());
    revenue2.setProductId('testProductId');
    assert.isTrue(revenue2._isValidRevenue());
  });

  it ('should convert into an object', function() {
    var productId = 'testProductId';
    var quantity = 15;
    var price = 10.99;
    var revenueType = 'testRevenueType'
    var properties = {'city': 'San Francisco'};

    var revenue = new Revenue().setProductId(productId).setQuantity(quantity).setPrice(price);
    revenue.setRevenueType(revenueType).setEventProperties(properties);

    var obj = revenue._toJSONObject();
    assert.deepEqual(obj, {
      '$productId': productId,
      '$quantity': quantity,
      '$price': price,
      '$revenueType': revenueType,
      'city': 'San Francisco'
    });
  });
});
