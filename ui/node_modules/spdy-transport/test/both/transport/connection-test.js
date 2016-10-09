var assert = require('assert');
var async = require('async');
var streamPair = require('stream-pair');
var fixtures = require('./fixtures');

var expectData = fixtures.expectData;
var everyProtocol = fixtures.everyProtocol;

var transport = require('../../../');

describe('Transport/Connection', function() {
  everyProtocol(function(name, version) {
    var server;
    var client;
    var pair;

    beforeEach(function() {
      server = fixtures.server;
      client = fixtures.client;
      pair = fixtures.pair;
    });

    it('should send SETTINGS frame on both ends', function(done) {
      async.map([ server, client ], function(side, callback) {
        side.on('frame', function(frame) {
          if (frame.type !== 'SETTINGS')
            return;

          callback();
        });
      }, done);
    });

    it('should emit `close` after GOAWAY', function(done) {
      client.request({
        path: '/hello-split'
      }, function(err, stream) {
        assert(!err);

        stream.resume();
        stream.end();
      });

      var once = false;
      server.on('stream', function(stream) {
        assert(!once);
        once = true;

        stream.respond(200, {});
        stream.resume();
        stream.end();

        var waiting = 2;
        function next() {
          if (--waiting === 0)
            done();
        }

        pair.destroySoon = next;
        server.once('close', next);
        server.end();
      });
    });

    it('should dump data on GOAWAY', function(done) {
      client.request({
        path: '/hello-split'
      }, function(err, stream) {
        assert(!err);

        stream.resume();
        stream.end();
      });

      var once = false;
      server.on('stream', function(stream) {
        assert(!once);
        once = true;

        stream.respond(200, {});
        stream.resume();
        stream.end();

        pair.destroySoon = function() {
          pair.end();
          server.ping();

          setTimeout(done, 10);
        };
        server.end();
      });
    });

    it('should kill late streams on GOAWAY', function(done) {
      client.request({
        path: '/hello-split'
      }, function(err, stream) {
        assert(!err);

        stream.resume();
        stream.end();

        client.request({
          path: '/late'
        }, function(err, stream) {
          assert(!err);

          stream.on('error', function() {
            done();
          });
        });
      });

      var once = false;
      server.on('stream', function(stream) {
        assert(!once);
        once = true;

        stream.respond(200, {});
        stream.resume();
        stream.end();

        server.end();
      });
    });

    it('should send and receive ping', function(done) {
      client.ping(function() {
        server.ping(done);
      });
    });

    it('should ignore request after GOAWAY', function(done) {
      client.request({
        path: '/hello-split'
      }, function(err, stream) {
        assert(!err);

        client.request({
          path: '/second'
        }, function(err, stream) {
          stream.on('error', function() {
            // Ignore
          });
        });
      });

      var once = false;
      server.on('stream', function(stream) {
        assert(!once);
        once = true;

        // Send GOAWAY
        server.end();
      });

      var waiting = 2;
      server.on('frame', function(frame) {
        if (frame.type === 'HEADERS' && --waiting === 0)
          setTimeout(done, 10);
      });
    });

    it('should return Stream after GOAWAY', function(done) {
      client.end(function() {
        var stream = client.request({
          path: '/hello-split'
        });
        assert(stream);

        stream.once('error', function() {
          done();
        });
      });
    });

    it('should timeout when sending request', function(done) {
      server.setTimeout(50, function() {
        server.end();
        setTimeout(done, 50);
      });

      setTimeout(function() {
        client.request({
          path: '/hello-with-data'
        }, function(err, stream) {
          assert(err);
        });
      }, 100);

      server.on('stream', function(stream) {
        assert(false);
      });
    });

    it('should not timeout when sending request', function(done) {
      server.setTimeout(100, function() {
        assert(false);
      });

      setTimeout(function() {
        client.request({
          path: '/hello-with-data'
        }, function(err, stream) {
          assert(!err);

          stream.end('ok');
          setTimeout(second, 50);
        });
      }, 50);

      function second() {
        client.request({
          path: '/hello-with-data'
        }, function(err, stream) {
          assert(!err);

          stream.end('ok');
          setTimeout(third, 50);
        });
      }

      function third() {
        client.ping(function() {
          server.end();
          setTimeout(done, 50);
        });
      }

      server.on('stream', function(stream) {
        stream.respond(200, {});
        stream.end();
        expectData(stream, 'ok', function() {});
      });
    });

    it('should ignore request without `stream` listener', function(done) {
      client.request({
        path: '/hello-split'
      }, function(err, stream) {
        assert(!err);

        stream.on('close', function(err) {
          assert(err);
          done();
        });
      });
    });

    it('should ignore HEADERS frame after FIN', function(done) {
      function sendHeaders() {
        client._spdyState.framer.requestFrame({
          id: 1,
          method: 'GET',
          path: '/',
          priority: null,
          headers: {},
          fin: true
        }, function(err) {
          assert(!err);
        });
      }

      var sent;
      client.request({
        path: '/hello'
      }, function(err, stream) {
        assert(!err);
        sent = true;

        stream.resume();
        stream.once('end', function() {
          stream.end(sendHeaders);
        });
      });

      var incoming = 0;
      server.on('stream', function(stream) {
        incoming++;
        assert(incoming <= 1);

        stream.resume();
        stream.end();
      });

      var waiting = 2;
      server.on('frame', function(frame) {
        if (frame.type === 'HEADERS' && --waiting === 0)
          process.nextTick(done);
      });
    });

    it('should use last received id when killing streams', function(done) {
      var waiting = 2;
      function next() {
        if (--waiting === 0)
          return done();
      }
      client.once('stream', next);
      server.once('stream', next);

      server.request({
        path: '/hello'
      }, function() {
        client.request({
          path: '/hello'
        });
      });
    });

    it('should kill stream on wrong id', function(done) {
      client._spdyState.stream.nextId = 2;

      var stream = client.request({
        path: '/hello'
      });
      stream.once('error', function(err) {
        done();
      });
    });

    it('should handle SETTINGS', function(done) {
      client._spdyState.framer.settingsFrame({
        max_frame_size: 100000,
        max_header_list_size: 1000,
        header_table_size: 32,
        enable_push: true
      }, function(err) {
        assert(!err);
      });
      client._spdyState.parser.setMaxFrameSize(100000);

      var sent;
      client.request({
        path: '/hello'
      }, function(err, stream) {
        assert(!err);
        sent = true;

        stream.on('data', function(chunk) {
          assert(chunk.length > 16384 || version < 4);
        });

        stream.once('end', done);
      });

      var incoming = 0;
      server.on('stream', function(stream) {
        incoming++;
        assert(incoming <= 1);

        stream.resume();
        server._spdyState.framer.dataFrame({
          id: stream.id,
          priority: stream._spdyState.priority.getPriority(),
          fin: true,
          data: new Buffer(32000)
        });
      });
    });

    it('should handle SETTINGS.initial_window_size=0', function(done) {
      var pair = streamPair.create();

      var client = transport.connection.create(pair.other, {
        protocol: name,
        windowSize: 256,
        isServer: false
      });
      client.start(version);

      var proto = transport.protocol[name];

      var framer = proto.framer.create({
        window: new transport.Window({
          id: 0,
          isServer: false,
          recv: { size: 1024 * 1024 },
          send: { size: 1024 * 1024 }
        })
      });
      var parser = proto.parser.create({
        window: new transport.Window({
          id: 0,
          isServer: false,
          recv: { size: 1024 * 1024 },
          send: { size: 1024 * 1024 }
        })
      });

      framer.setVersion(version);
      parser.setVersion(version);

      var pool = proto.compressionPool.create();
      var comp = pool.get(version);
      framer.setCompression(comp);
      parser.setCompression(comp);

      framer.pipe(pair);
      pair.pipe(parser);

      framer.settingsFrame({
        initial_window_size: 0
      }, function(err) {
        assert(!err);
      });

      client.on('frame', function(frame) {
        if (frame.type !== 'SETTINGS')
          return;

        client.request({
          path: '/hello'
        }, function(err, stream) {
          assert(!err);

          // Attempt to get data through
          setTimeout(done, 100);
        }).end('hello');
      });

      parser.on('data', function(frame) {
        assert.notEqual(frame.type, 'DATA');
      });
    });

    if (version >= 4) {
      it('should ignore too large HPACK table in SETTINGS', function(done) {
        var limit = 0xffffffff;
        server._spdyState.framer.settingsFrame({
          header_table_size: limit
        }, function(err) {
          assert(!err);
        });

        var headers = {};
        for (var i = 0; i < 2048; i++)
          headers['h' + i] = (i % 250).toString();

        client.on('frame', function(frame) {
          if (frame.type !== 'SETTINGS' ||
              frame.settings.header_table_size !== 0xffffffff) {
            return;
          }

          // Time for request!
          var one = client.request({
            headers: headers,
            path: '/hello'
          });
          one.end();
          one.resume();
        });

        server.on('frame', function(frame) {
          if (frame.type === 'SETTINGS') {
            // Emulate bigger table on server-side
            server._spdyState.pair.decompress._table.protocolMaxSize = limit;
            server._spdyState.pair.decompress._table.maxSize = limit;
          }

          if (frame.type !== 'HEADERS')
            return;

          assert.equal(server._spdyState.pair.decompress._table.size, 4062);
          assert.equal(client._spdyState.pair.compress._table.size, 4062);
          assert.equal(client._spdyState.pair.compress._table.maxSize,
                       client._spdyState.constants.HEADER_TABLE_SIZE);
          done();
        });
      });

      it('should allow receiving PRIORITY on idle stream', function(done) {
        client._spdyState.framer.priorityFrame({
          id: 5,
          priority: {
            exclusive: false,
            parent: 3,
            weight: 10
          }
        }, function() {
        });

        server.on('frame', function(frame) {
          if (frame.type === 'PRIORITY') {
            setImmediate(done);
          }
        });

        client.on('frame', function(frame) {
          assert.notEqual(frame.type, 'GOAWAY');
        });
      });

      it('should allow receiving PRIORITY on small-id stream', function(done) {
        server.on('stream', function(stream) {
          stream.end();
        });

        client._spdyState.stream.nextId = 3;

        var one = client.request({
          path: '/hello'
        });
        one.end();
        one.resume();

        one.on('close', function() {
          client._spdyState.framer.priorityFrame({
            id: 1,
            priority: {
              exclusive: false,
              parent: 3,
              weight: 10
            }
          }, function() {
          });
        });

        server.on('frame', function(frame) {
          if (frame.type === 'PRIORITY' && frame.id === 1) {
            setImmediate(done);
          }
        });

        client.removeAllListeners('frame');
        client.on('frame', function(frame) {
          assert.notEqual(frame.type, 'GOAWAY');
        });
      });

      it('should allow receiving PRIORITY on even-id stream', function(done) {
        client._spdyState.framer.priorityFrame({
          id: 2,
          priority: {
            exclusive: false,
            parent: 3,
            weight: 10
          }
        }, function() {
        });

        server.on('frame', function(frame) {
          if (frame.type === 'PRIORITY' && frame.id === 2) {
            setImmediate(done);
          }
        });

        client.removeAllListeners('frame');
        client.on('frame', function(frame) {
          assert.notEqual(frame.type, 'GOAWAY');
        });
      });
    }

    it('should send X_FORWARDED_FOR', function(done) {
      client.sendXForwardedFor('1.2.3.4');

      var sent;
      client.request({
        path: '/hello'
      }, function(err, stream) {
        assert(!err);
        sent = true;

        stream.resume();
        stream.once('end', done);
      });

      server.on('stream', function(stream) {
        assert.equal(server.getXForwardedFor(), '1.2.3.4');

        stream.resume();
        stream.end();
      });
    });
  });
});
