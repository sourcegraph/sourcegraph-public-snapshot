var UUID = require('../src/uuid.js');

describe('UUID', function() {
  var encodeCases = [
    ['', 'd41d8cd98f00b204e9800998ecf8427e'],
    ['foobar', '3858f62230ac3c915f300c664312c63f']
  ];

  it('should generate a valid UUID-4', function() {
    var uuid = UUID();
    assert.equal(36, uuid.length);
    assert.equal('4', uuid.substring(14, 15));
  });
});
