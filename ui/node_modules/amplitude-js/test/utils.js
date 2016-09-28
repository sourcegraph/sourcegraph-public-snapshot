describe('utils', function() {
  var utils = require('../src/utils.js');

  describe('isEmptyString', function() {
    it('should detect empty strings', function() {
      assert.isTrue(utils.isEmptyString(null));
      assert.isTrue(utils.isEmptyString(''));
      assert.isTrue(utils.isEmptyString(undefined));
      assert.isTrue(utils.isEmptyString(NaN));
      assert.isFalse(utils.isEmptyString(' '));
      assert.isFalse(utils.isEmptyString('string'));
      assert.isFalse(utils.isEmptyString("string"));
    });
  });

  describe('validateProperties', function() {
    it('should detect invalid event property formats', function() {
      assert.deepEqual({}, utils.validateProperties('string'));
      assert.deepEqual({}, utils.validateProperties(null));
      assert.deepEqual({}, utils.validateProperties(undefined));
      assert.deepEqual({}, utils.validateProperties(10));
      assert.deepEqual({}, utils.validateProperties(true));
      assert.deepEqual({}, utils.validateProperties(new Date()));
      assert.deepEqual({}, utils.validateProperties([]));
      assert.deepEqual({}, utils.validateProperties(NaN));
    });

    it('should not modify valid event property formats', function() {
      var properties = {
        'test': 'yes',
        'key': 'value',
        '15': '16'
      }
      assert.deepEqual(properties, utils.validateProperties(properties));
    });

    it('should coerce non-string keys', function() {
      var d = new Date();
      var dateString = String(d);

      var properties = {
        10: 'false',
        null: 'value',
        NaN: '16',
        d: dateString
      }
      var expected = {
        '10': 'false',
        'null': 'value',
        'NaN': '16',
        'd': dateString
      }
      assert.deepEqual(utils.validateProperties(properties), expected);
    });

    it('should ignore invalid event property values', function() {
      var properties = {
        'null': null,
        'undefined': undefined,
        'NaN': NaN,
        'function': utils.log
      }
      assert.deepEqual({}, utils.validateProperties(properties));
    });

    it('should coerce error values', function() {
      var e = new Error('oops');

      var properties = {
        'error': e
      };
      var expected = {
        'error': String(e)
      }
      assert.deepEqual(utils.validateProperties(properties), expected);
    });

    it('should validate properties', function() {
      var e = new Error('oops');

      var properties = {
        10: 'false', // coerce key
        'bool': true,
        'null': null, // should be ignored
        'function': console.log, // should be ignored
        'regex': /afdg/, // should be ignored
        'error': e, // coerce value
        'string': 'test',
        'array': [0, 1, 2, '3'],
        'nested_array': ['a', {'key': 'value'}, ['b']],
        'object': {
          'key': 'value',
          15: e
        },
        'nested_object': {
          'k': 'v',
          'l': [0, 1],
          'o': {
              'k2': 'v2',
              'l2': ['e2', {'k3': 'v3'}]
          }
        }
      }
      var expected = {
        '10': 'false',
        'bool': true,
        'error': 'Error: oops',
        'string': 'test',
        'array': [0, 1, 2, '3'],
        'nested_array': ['a'],
        'object': {
          'key': 'value',
          '15': 'Error: oops'
        },
        'nested_object': {
          'k': 'v',
          'l': [0, 1],
          'o': {
              'k2': 'v2',
              'l2': ['e2']
          }
        }
      }
      assert.deepEqual(utils.validateProperties(properties), expected);
    });
  });
});
