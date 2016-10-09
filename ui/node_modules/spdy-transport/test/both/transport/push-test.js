var assert = require('assert');
var async = require('async');
var fixtures = require('./fixtures');

var expectData = fixtures.expectData;
var everyProtocol = fixtures.everyProtocol;

var transport = require('../../../');

describe('Transport/Push', function() {
  everyProtocol(function(name, version) {
    var server;
    var client;
    var pair;

    beforeEach(function() {
      server = fixtures.server;
      client = fixtures.client;
      pair = fixtures.pair;
    });

    it('should create PUSH_PROMISE', function(done) {
      client.request({
        path: '/parent'
      }, function(err, stream) {
        assert(!err);

        stream.on('pushPromise', function(push) {
          assert.equal(push.path, '/push');
          assert.equal(client.getCounter('push'), 1);
          push.on('response', function(status, headers) {
            assert.equal(status, 201);
            done();
          });
        });
      });

      server.on('stream', function(stream) {
        assert.equal(stream.path, '/parent');

        stream.respond(200, {});
        stream.pushPromise({
          path: '/push',
          status: 201,
          priority: {
            parent: 0,
            exclusive: false,
            weight: 42
          }
        }, function(err, stream) {
          assert(!err);
        });
      });
    });

    it('should send HEADERS on PUSH_PROMISE', function(done) {
      client.request({
        path: '/parent'
      }, function(err, stream) {
        assert(!err);

        stream.on('pushPromise', function(push) {
          push.on('headers', function(headers) {
            assert.deepEqual(headers, { a: 'b' });
            done();
          });
        });
      });

      server.on('stream', function(stream) {
        assert.equal(stream.path, '/parent');

        stream.respond(200, {});
        stream.pushPromise({
          path: '/push',
          priority: {
            parent: 0,
            exclusive: false,
            weight: 42
          }
        }, function(err, stream) {
          assert(!err);

          stream.sendHeaders({ a: 'b' });
        });
      });
    });

    if (version >= 4) {
      it('should send PUSH_PROMISE+HEADERS and HEADERS concurrently',
         function(done) {
        var seq = [];

        client.request({
          path: '/parent'
        }, function(err, stream) {
          assert(!err);

          stream.on('pushPromise', function(push) {
            assert.equal(push.path, '/push');
            assert.equal(client.getCounter('push'), 1);
            push.on('response', function(status) {
              assert.equal(status, 201);
              assert.deepEqual(seq, [ 0, 1, 2 ]);
              done();
            });
          });
        });

        client.on('frame', function(frame) {
          if (frame.type === 'HEADERS' || frame.type === 'PUSH_PROMISE')
            seq.push(frame.headers.seq | 0);
        });

        server.on('stream', function(stream) {
          assert.equal(stream.path, '/parent');

          stream.pushPromise({
            path: '/push',
            status: 201,
            priority: {
              parent: 0,
              exclusive: false,
              weight: 42
            },
            headers: {
              'seq': '0'
            },
            response: {
              'seq': '2'
            }
          }, function(err, stream) {
            assert(!err);
          });
          process.nextTick(function() {
            stream.respond(200, {
              'seq': '1'
            });
          });
        });
      });
    }

    it('should create PUSH_PROMISE and end parent req', function(done) {
      client.request({
        path: '/parent'
      }, function(err, stream) {
        assert(!err);

        stream.resume();
        stream.end();
        stream.on('pushPromise', function(push) {
          assert.equal(push.path, '/push');
          done();
        });
      });

      server.on('stream', function(stream) {
        assert.equal(stream.path, '/parent');

        stream.respond(200, {});
        stream.resume();
        stream.on('end', function() {
          stream.pushPromise({
            path: '/push',
            priority: {
              parent: 0,
              exclusive: false,
              weight: 42
            }
          }, function(err, stream) {
            assert(!err);
          });
          stream.end();
        });
      });
    });

    it('should cork PUSH_PROMISE on write', function(done) {
      client.request({
        path: '/parent'
      }, function(err, stream) {
        assert(!err);

        stream.on('pushPromise', function(push) {
          assert.equal(push.path, '/push');
          expectData(push, 'ok', done);
        });
      });

      server.on('stream', function(stream) {
        assert.equal(stream.path, '/parent');

        stream.respond(200, {});
        var push = stream.pushPromise({
          path: '/push',
          priority: {
            parent: 0,
            exclusive: false,
            weight: 42
          }
        }, function(err, stream) {
          assert(!err);
        });

        push.end('ok');
      });
    });

    it('should emit `close` on PUSH_PROMISE', function(done) {
      client.request({
        path: '/parent'
      }, function(err, stream) {
        assert(!err);

        stream.on('pushPromise', function(push) {
          assert.equal(push.path, '/push');

          push.on('close', next);
          push.resume();
        });
      });

      server.on('stream', function(stream) {
        assert.equal(stream.path, '/parent');

        stream.respond(200, {});
        stream.pushPromise({
          path: '/push',
          priority: {
            parent: 0,
            exclusive: false,
            weight: 42
          }
        }, function(err, stream) {
          assert(!err);
          stream.on('close', next);
          stream.end('ohai');
        });
      });

      var waiting = 2;
      function next() {
        if (--waiting === 0)
          return done();
      }
    });

    it('should ignore PUSH_PROMISE', function(done) {
      client.request({
        path: '/parent'
      }, function(err, stream) {
        assert(!err);
      });

      server.on('stream', function(stream) {
        assert.equal(stream.path, '/parent');

        stream.respond(200, {});
        stream.pushPromise({
          path: '/push',
          priority: {
            parent: 0,
            exclusive: false,
            weight: 42
          }
        }, function(err, stream) {
          assert(!err);
          stream.once('close', function(err) {
            assert(err);
            done();
          });
        });
      });
    });

    it('should fail on server-disabled PUSH_PROMISE', function(done) {
      client.request({
        path: '/parent'
      }, function(err, stream) {
        assert(!err);

        stream._spdyState.framer.enablePush(true);
        stream.pushPromise({
          path: '/push',
          priority: {
            parent: 0,
            exclusive: false,
            weight: 42
          }
        }, function(err, stream) {
          assert(!err);
          stream.on('error', function(err) {
            assert(err);
          });
        });

        client.on('close', function(err) {
          assert(err);
          done();
        });
      });

      server.on('stream', function(stream) {
        assert.equal(stream.path, '/parent');

        stream.respond(200, {});
        stream.on('pushPromise', function() {
          assert(false);
        });
      });
    });

    it('should fail on client-disabled PUSH_PROMISE', function(done) {
      client.request({
        path: '/parent'
      }, function(err, stream) {
        assert(!err);

        stream._spdyState.framer.enablePush(false);
        var push = stream.pushPromise({
          path: '/push',
          priority: {
            parent: 0,
            exclusive: false,
            weight: 42
          }
        }, function(err, stream) {
          assert(err);
          setTimeout(function() {
            done();
          }, 50);
        });
        push.write('hello');
      });

      // The PUSH data should not be sent
      server.on('frame', function(frame) {
        assert.notEqual(frame.type, 'DATA');
      });

      server.on('stream', function(stream) {
        assert.equal(stream.path, '/parent');

        stream.respond(200, {});
        stream.on('pushPromise', function() {
          assert(false);
        });
      });
    });

    it('should get error on disabled PUSH_PROMISE', function(done) {
      client.request({
        path: '/parent'
      }, function(err, stream) {
        assert(!err);

        stream.pushPromise({
          path: '/push',
          priority: {
            parent: 0,
            exclusive: false,
            weight: 42
          }
        }, function(err, stream) {
          assert(err);
          done();
        });
      });

      server.on('stream', function(stream) {
        assert.equal(stream.path, '/parent');

        stream.respond(200, {});

        stream.on('pushPromise', function() {
          assert(false);
        });
      });
    });

    it('should not error on extra PRIORITY frame', function(done) {
      client.request({
        path: '/parent'
      }, function(err, stream) {
        assert(!err);

        stream.on('pushPromise', function(push) {
          push.on('response', function() {
            // .abort() does this only on next tick
            push.emit('close');

            stream.end('ok');
          });
        });
      });

      server.on('stream', function(stream) {
        assert.equal(stream.path, '/parent');

        stream.respond(200, {});
        stream.pushPromise({
          path: '/push',
          priority: {
            parent: 0,
            exclusive: false,
            weight: 42
          }
        }, function(err, stream) {
          assert(!err);
          stream.on('error', function() {
            assert(false);
          });
        });

        expectData(stream, 'ok', done);
      });
    });

    it('should not fail on aborting PUSH_PROMISE frame', function(done) {
      client.request({
        path: '/parent'
      }, function(err, stream) {
        assert(!err);

        stream.on('pushPromise', function(push) {
          push.abort();
          stream.end('ok');
        });
      });

      server.on('stream', function(stream) {
        assert.equal(stream.path, '/parent');

        stream.respond(200, {});
        stream.pushPromise({
          path: '/push',
          priority: {
            parent: 0,
            exclusive: false,
            weight: 42
          }
        }, function(err, stream) {
          assert(!err);
          stream.on('error', function() {
            assert(false);
          });
        });

        expectData(stream, 'ok', done);
      });
    });
  });
});
