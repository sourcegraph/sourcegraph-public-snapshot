var Identify = require('../src/identify.js');

describe('Identify', function() {

  it('should unset properties', function () {
    var property1 = 'testProperty1';
    var property2 = 'testProperty2';
    var identify = new Identify().unset(property1).unset(property2);

    var expected = {
      '$unset': {}
    };
    expected['$unset'][property1] = '-';
    expected['$unset'][property2] = '-';

    assert.deepEqual(expected, identify.userPropertiesOperations);
    assert.deepEqual([property1, property2], identify.properties);
  });

  it('should set properties', function () {
    var property1 = 'var value';
    var value1 = 'testValue';

    var property2 = 'float value';
    var value2 = 0.123;

    var property3 = 'bool value';
    var value3 = true;

    var property4 = 'json value';
    var value4 = {};

    var identify = new Identify().set(property1, value1).set(property2, value2);
    identify.set(property3, value3).set(property4, value4);

    // identify should ignore this since duplicate key
    identify.set(property1, value3);

    var expected = {
      '$set': {}
    }
    expected['$set'][property1] = value1;
    expected['$set'][property2] = value2;
    expected['$set'][property3] = value3;
    expected['$set'][property4] = value4;

    assert.deepEqual(expected, identify.userPropertiesOperations);
    assert.deepEqual([property1, property2, property3, property4], identify.properties);
  });

  it ('should set properties once', function () {
    var property1 = 'var value';
    var value1 = 'testValue';

    var property2 = 'float value';
    var value2 = 0.123;

    var property3 = 'bool value';
    var value3 = true;

    var property4 = 'json value';
    var value4 = {};

    var identify = new Identify().setOnce(property1, value1).setOnce(property2, value2);
    identify.setOnce(property3, value3).setOnce(property4, value4);

    // identify should ignore this since duplicate key
    identify.setOnce(property1, value3);

    var expected = {
      '$setOnce': {}
    }
    expected['$setOnce'][property1] = value1;
    expected['$setOnce'][property2] = value2;
    expected['$setOnce'][property3] = value3;
    expected['$setOnce'][property4] = value4;

    assert.deepEqual(expected, identify.userPropertiesOperations);
    assert.deepEqual([property1, property2, property3, property4], identify.properties);
  });

  it ('should add properties', function () {
    var property1 = 'int value';
    var bad_value1 = ['test', 'array']; // add should filter out arrays
    var value1 = 5;

    var property2 = 'var value';
    var bad_value2 = { // add should filter out maps
      'test': 'array'
    };
    var value2 = 0.123;

    var identify = new Identify().add(property1, bad_value1).add(property1, value1);
    identify.add(property2, bad_value2).add(property2, value2);

    // identify should ignore this since duplicate key
    identify.add(property1, 'duplicate');

    var expected = {
      '$add': {}
    };
    expected['$add'][property1] = value1;
    expected['$add'][property2] = value2;

    assert.deepEqual(expected, identify.userPropertiesOperations);
    assert.deepEqual([property1, property2], identify.properties);
  });

  it ('should append properties', function () {
    var property1 = 'var value';
    var value1 = 'testValue';

    var property2 = 'float value';
    var value2 = 0.123;

    var property3 = 'bool value';
    var value3 = true;

    var property4 = 'json value';
    var value4 = {};

    var property5 = 'list value';
    var value5 = [1, 2, 'test'];

    var identify = new Identify().append(property1, value1).append(property2, value2);
    identify.append(property3, value3).append(property4, value4).append(property5, value5);

    // identify should ignore this since duplicate key
    identify.setOnce(property1, value3);

    var expected = {
      '$append': {}
    }
    expected['$append'][property1] = value1;
    expected['$append'][property2] = value2;
    expected['$append'][property3] = value3;
    expected['$append'][property4] = value4;
    expected['$append'][property5] = value5;

    assert.deepEqual(expected, identify.userPropertiesOperations);
    assert.deepEqual([property1, property2, property3, property4, property5], identify.properties);
  });

    it ('should prepend properties', function () {
    var property1 = 'var value';
    var value1 = 'testValue';

    var property2 = 'float value';
    var value2 = 0.123;

    var property3 = 'bool value';
    var value3 = true;

    var property4 = 'json value';
    var value4 = {};

    var property5 = 'list value';
    var value5 = [1, 2, 'test'];

    var identify = new Identify().prepend(property1, value1).prepend(property2, value2);
    identify.prepend(property3, value3).prepend(property4, value4).prepend(property5, value5);

    // identify should ignore this since duplicate key
    identify.setOnce(property1, value3);

    var expected = {
      '$prepend': {}
    }
    expected['$prepend'][property1] = value1;
    expected['$prepend'][property2] = value2;
    expected['$prepend'][property3] = value3;
    expected['$prepend'][property4] = value4;
    expected['$prepend'][property5] = value5;

    assert.deepEqual(expected, identify.userPropertiesOperations);
    assert.deepEqual([property1, property2, property3, property4, property5], identify.properties);
  });

  it ('should allow multiple operations', function () {
    var property1 = 'string value';
    var value1 = 'testValue';

    var property2 = 'float value';
    var value2 = 0.123;

    var property3 = 'bool value';
    var value3 = true;

    var property4 = 'json value';

    var property5 = 'list value';
    var value5 = [1, 2, 'test'];

    var property6 = 'int value';
    var value6 = 100;

    var identify = new Identify().setOnce(property1, value1).add(property2, value2);
    identify.set(property3, value3).unset(property4).append(property5, value5);
    identify.prepend(property6, value6);

    // identify should ignore this since duplicate key
    identify.set(property4, value3);

    var expected = {
      '$add': {},
      '$append': {},
      '$prepend': {},
      '$set': {},
      '$setOnce': {},
      '$unset': {}
    };
    expected['$setOnce'][property1] = value1;
    expected['$add'][property2] = value2;
    expected['$set'][property3] = value3;
    expected['$unset'][property4] = '-';
    expected['$append'][property5] = value5;
    expected['$prepend'][property6] = value6;

    assert.deepEqual(expected, identify.userPropertiesOperations);
    assert.deepEqual([property1, property2, property3, property4, property5, property6], identify.properties);
  });

  it ('should disallow duplicate properties', function () {
    var property = "testProperty";
    var value1 = "testValue";
    var value2 = 0.123;
    var value3 = true;
    var value4 = {};

    var identify = new Identify().setOnce(property, value1).add(property, value2);
    identify.set(property, value3).unset(property);

    var expected = {
      '$setOnce': {}
    };
    expected['$setOnce'][property] = value1

    assert.deepEqual(expected, identify.userPropertiesOperations);
    assert.deepEqual([property], identify.properties);
  });

  it ('should disallow other operations on a clearAll identify', function() {
    var property = "testProperty";
    var value1 = "testValue";
    var value2 = 0.123;
    var value3 = true;
    var value4 = {};

    var identify = new Identify().clearAll();
    identify.setOnce(property, value1).add(property, value2).set(property, value3).unset(property);

    var expected = {
        '$clearAll': '-'
    };

    assert.deepEqual(expected, identify.userPropertiesOperations);
    assert.deepEqual([], identify.properties);
  });

  it ('should disallow clearAll on an identify with other operations', function() {
    var property = "testProperty";
    var value1 = "testValue";
    var value2 = 0.123;
    var value3 = true;
    var value4 = {};

    var identify = new Identify().setOnce(property, value1).add(property, value2);
    identify.set(property, value3).unset(property).clearAll();

    var expected = {
      '$setOnce': {}
    };
    expected['$setOnce'][property] = value1

    assert.deepEqual(expected, identify.userPropertiesOperations);
    assert.deepEqual([property], identify.properties);
  });

  it ('should not log any warnings for calling clearAll multiple times on a single identify', function() {
    var identify = new Identify().clearAll().clearAll().clearAll();
  });
});
