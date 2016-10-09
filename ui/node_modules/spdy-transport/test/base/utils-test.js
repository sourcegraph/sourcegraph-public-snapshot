var assert = require('assert');

var transport = require('../../');
var utils = transport.utils;

describe('utils', function() {
  function compare(a, b) {
    return a - b;
  }

  describe('binaryInsert', function() {
    var binaryInsert = utils.binaryInsert;
    it('should properly insert items in sequential order', function() {
      var list = [];
      binaryInsert(list, 1, compare);
      binaryInsert(list, 2, compare);
      binaryInsert(list, 3, compare);
      binaryInsert(list, 4, compare);

      assert.deepEqual(list, [ 1, 2, 3, 4 ]);
    });

    it('should properly insert items in reverse order', function() {
      var list = [];
      binaryInsert(list, 4, compare);
      binaryInsert(list, 3, compare);
      binaryInsert(list, 2, compare);
      binaryInsert(list, 1, compare);

      assert.deepEqual(list, [ 1, 2, 3, 4 ]);
    });

    it('should properly insert items in random order', function() {
      var list = [];
      binaryInsert(list, 3, compare);
      binaryInsert(list, 2, compare);
      binaryInsert(list, 4, compare);
      binaryInsert(list, 1, compare);

      assert.deepEqual(list, [ 1, 2, 3, 4 ]);
    });
  });

  describe('binarySearch', function() {
    var binarySearch = utils.binarySearch;

    it('should return the index of the value', function() {
      var list = [ 1, 2, 3, 4, 5, 6, 7 ];
      for (var i = 0; i < list.length; i++)
        assert.equal(binarySearch(list, list[i], compare), i);
    });

    it('should return -1 when value is not present in list', function() {
      var list = [ 1, 2, 3, 5, 6, 7 ];
      assert.equal(binarySearch(list, 4, compare), -1);
      assert.equal(binarySearch(list, 0, compare), -1);
      assert.equal(binarySearch(list, 8, compare), -1);
    });
  });

  describe('priority to weight', function() {
    var utils = transport.protocol.base.utils;

    var toWeight = utils.priorityToWeight;
    var toPriority = utils.weightToPriority;

    it('should preserve weight=16', function() {
      var priority = toPriority(16);
      assert.equal(priority, 3);
      assert.equal(toWeight(priority), 16);
    });
  });
});
