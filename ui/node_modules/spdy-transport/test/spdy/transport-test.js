var assert = require('assert');
var async = require('async');
var streamPair = require('stream-pair');

var transport = require('../../');

describe('SPDY Transport', function() {
  var server = null;
  var client = null;

  beforeEach(function() {
    var pair = streamPair.create();

    server = transport.connection.create(pair, {
      protocol: 'spdy',
      windowSize: 256,
      isServer: true,
      autoSpdy31: true
    });

    client = transport.connection.create(pair.other, {
      protocol: 'spdy',
      windowSize: 256,
      isServer: false,
      autoSpdy31: true
    });
  });

  describe('autoSpdy31', function() {
    it('should automatically switch on server', function(done) {
      server.start(3);
      assert.equal(server.getVersion(), 3);

      client.start(3.1);

      server.on('version', function() {
        assert.equal(server.getVersion(), 3.1);
        done();
      });
    });
  });

  describe('version detection', function() {
    it('should detect v2 on server', function(done) {
      client.start(2);

      server.on('version', function() {
        assert.equal(server.getVersion(), 2);
        done();
      });
    });

    it('should detect v3 on server', function(done) {
      client.start(3);

      server.on('version', function() {
        assert.equal(server.getVersion(), 3);
        done();
      });
    });
  });

  it('it should not wait for id=0 WINDOW_UPDATE on v3', function(done) {
    client.start(3);

    var buf = new Buffer(64 * 1024);
    buf.fill('x');

    client.request({
      method: 'POST',
      path: '/',
      headers: {}
    }, function(err, stream) {
      assert(!err);

      stream.write(buf);
      stream.write(buf);
      stream.write(buf);
      stream.end(buf);
    });

    server.on('stream', function(stream) {
      stream.respond(200, {});

      var received = 0;
      stream.on('data', function(chunk) {
        received += chunk.length;
      });

      stream.on('end', function() {
        assert.equal(received, buf.length * 4);
        done();
      });
    });
  });
});
