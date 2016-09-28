var Base64 = require('../src/base64.js');

describe('Base64', function() {
  var encodeCases = [
    ['', ''],
    ['f', 'Zg=='],
    ['fo', 'Zm8='],
    ['foo', 'Zm9v'],
    ['foob', 'Zm9vYg=='],
    ['fooba', 'Zm9vYmE='],
    ['foobar', 'Zm9vYmFy'],
    ['\u2661', '4pmh']
  ];

  var decodeCases = [
    ['', ''],
    ['Zg', 'f'],
    ['Zg==', 'f'],
    ['Zm8', 'fo'],
    ['Zm8=', 'fo'],
    ['Zm9v', 'foo'],
    ['Zm9vYg', 'foob'],
    ['Zm9vYg==', 'foob'],
    ['Zm9vYmE', 'fooba'],
    ['Zm9vYmE=', 'fooba'],
    ['Zm9vYmFy', 'foobar'],
    ['4pmh', '\u2661']
  ];

  it('should encode properly', function() {
    for (var i = 0; i < encodeCases.length; i++) {
      assert.equal(encodeCases[i][1], Base64.encode(encodeCases[i][0]));
    }
  });

  it('should decode properly', function() {
    for (var i = 0; i < decodeCases.length; i++) {
      assert.equal(decodeCases[i][1], Base64.decode(decodeCases[i][0]));
    }
  });
});
