var assert = require('assert');
var streamPair = require('stream-pair');

var transport = require('../../../');

exports.pair = null;
exports.server = null;
exports.client = null;

function expectData(stream, expected, callback) {
  var actual = '';

  stream.on('data', function(chunk) {
    actual += chunk;
  });
  stream.on('end', function() {
    assert.equal(actual, expected);
    callback();
  });
}
exports.expectData = expectData;

function protocol(name, version, body) {
  describe(name + ' (v' + version + ')', function() {
    beforeEach(function() {
      exports.pair = streamPair.create();

      exports.server = transport.connection.create(exports.pair, {
        protocol: name,
        windowSize: 256,
        isServer: true
      });
      exports.client = transport.connection.create(exports.pair.other, {
        protocol: name,
        windowSize: 256,
        isServer: false
      });

      exports.client.start(version);
    });

    body(name, version);
  });
}
exports.protocol = protocol;

function everyProtocol(body) {
  protocol('http2', 4, body);
  protocol('spdy', 2, body);
  protocol('spdy', 3, body);
  protocol('spdy', 3.1, body);
}
exports.everyProtocol = everyProtocol;
