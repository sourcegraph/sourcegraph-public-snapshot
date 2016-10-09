var assert = require('assert');

var transport = require('../../');

describe('Framer', function() {
  var framer;
  var parser;

  function protocol(name, version, body) {
    describe(name + ' (v' + version + ')', function() {
      beforeEach(function() {
        var proto = transport.protocol[name];

        var pool = proto.compressionPool.create();
        framer = proto.framer.create({
          window: new transport.Window({
            id: 0,
            isServer: false,
            recv: { size: 1024 * 1024 },
            send: { size: 1024 * 1024 }
          })
        });
        parser = proto.parser.create({
          isServer: true,
          window: new transport.Window({
            id: 0,
            isServer: true,
            recv: { size: 1024 * 1024 },
            send: { size: 1024 * 1024 }
          })
        });

        var comp = pool.get(version);
        framer.setCompression(comp);
        parser.setCompression(comp);

        framer.setVersion(version);
        parser.setVersion(version);

        parser.skipPreface();

        framer.pipe(parser);
      });

      body(name, version);
    });
  }

  function everyProtocol(body) {
    protocol('http2', 4, body);
    protocol('spdy', 2, body);
    protocol('spdy', 3, body);
    protocol('spdy', 3.1, body);
  }

  function expect(expected, done) {
    var acc = [];
    if (!Array.isArray(expected))
      expected = [ expected ];
    parser.on('data', function(frame) {
      acc.push(frame);

      if (acc.length !== expected.length)
        return;

      assert.deepEqual(acc, expected);
      done();
    });
  }

  everyProtocol(function(name, version) {
    describe('SETTINGS', function() {
      it('should generate empty frame', function(done) {
        framer.settingsFrame({}, function(err) {
          assert(!err);

          expect({
            type: 'SETTINGS',
            settings: {}
          }, done);
        });
      });

      it('should generate regular frame', function(done) {
        framer.settingsFrame({
          max_concurrent_streams: 100,
          initial_window_size: 42
        }, function(err) {
          assert(!err);

          expect({
            type: 'SETTINGS',
            settings: {
              max_concurrent_streams: 100,
              initial_window_size: 42
            }
          }, done);
        });
      });

      it('should not put Infinity values', function(done) {
        framer.settingsFrame({
          max_concurrent_streams: Infinity
        }, function(err) {
          assert(!err);

          expect({
            type: 'SETTINGS',
            settings: {}
          }, done);
        });
      });

      if (version >= 4) {
        it('should generate ACK frame', function(done) {
          framer.ackSettingsFrame(function(err) {
            assert(!err);

            expect({
              type: 'ACK_SETTINGS'
            }, done);
          });
        });
      }
    });

    describe('WINDOW_UPDATE', function() {
      it('should generate regular frame', function(done) {
        framer.windowUpdateFrame({
          id: 41,
          delta: 257
        }, function(err) {
          assert(!err);

          expect({
            type: 'WINDOW_UPDATE',
            id: 41,
            delta: 257
          }, done);
        });
      });

      it('should generate negative delta frame', function(done) {
        framer.windowUpdateFrame({
          id: 41,
          delta: -257
        }, function(err) {
          assert(!err);

          expect({
            type: 'WINDOW_UPDATE',
            id: 41,
            delta: -257
          }, done);
        });
      });
    });

    describe('DATA', function() {
      it('should generate regular frame', function(done) {
        framer.dataFrame({
          id: 41,
          priority: 0,
          fin: false,
          data: new Buffer('hello')
        }, function(err) {
          assert(!err);

          expect({
            type: 'DATA',
            id: 41,
            fin: false,
            data: new Buffer('hello')
          }, done);
        });
      });

      it('should generate fin frame', function(done) {
        framer.dataFrame({
          id: 41,
          priority: 0,
          fin: true,
          data: new Buffer('hello')
        }, function(err) {
          assert(!err);

          expect({
            type: 'DATA',
            id: 41,
            fin: true,
            data: new Buffer('hello')
          }, done);
        });
      });

      it('should generate empty frame', function(done) {
        framer.dataFrame({
          id: 41,
          priority: 0,
          fin: false,
          data: new Buffer(0)
        }, function(err) {
          assert(!err);

          expect({
            type: 'DATA',
            id: 41,
            fin: false,
            data: new Buffer(0)
          }, done);
        });
      });

      it('should split frame in multiple', function(done) {
        framer.setMaxFrameSize(10);
        parser.setMaxFrameSize(10);

        var big = new Buffer(32);
        big.fill('A');

        framer.dataFrame({
          id: 41,
          priority: 0,
          fin: false,
          data: big
        }, function(err) {
          assert(!err);

          var waiting = big.length;
          var actual = '';
          parser.on('data', function(frame) {
            assert.equal(frame.type, 'DATA');
            actual += frame.data;
            waiting -= frame.data.length;
            if (waiting !== 0)
              return;

            assert.equal(actual, big.toString());
            done();
          });
        });
      });

      it('should update window on both sides', function(done) {
        framer.dataFrame({
          id: 41,
          priority: 0,
          fin: false,
          data: new Buffer('hello')
        }, function(err) {
          assert(!err);

          expect({
            type: 'DATA',
            id: 41,
            fin: false,
            data: new Buffer('hello')
          }, function() {
            assert.equal(framer.window.send.current,
                         parser.window.recv.current);
            assert.equal(framer.window.send.current, 1024 * 1024 - 5);
            done();
          });
        });
      });
    });

    describe('HEADERS', function() {
      it('should generate request frame', function(done) {
        framer.requestFrame({
          id: 1,
          path: '/',
          host: 'localhost',
          method: 'GET',
          headers: {
            a: 'b',
            host: 'localhost',

            // Should be removed
            connection: 'keep-alive'
          }
        }, function(err) {
          assert(!err);

          expect({
            type: 'HEADERS',
            id: 1,
            fin: false,
            writable: true,
            priority: {
              weight: 16,
              parent: 0,
              exclusive: false
            },
            path: '/',
            headers: {
              ':authority': 'localhost',
              ':path': '/',
              ':scheme': 'https',
              ':method': 'GET',

              a: 'b'
            }
          }, done);
        });
      });

      it('should skip internal headers', function(done) {
        framer.requestFrame({
          id: 1,
          path: '/',
          host: 'localhost',
          method: 'GET',
          headers: {
            a: 'b',
            host: 'localhost',
            ':method': 'oopsie'
          }
        }, function(err) {
          assert(!err);

          expect({
            type: 'HEADERS',
            id: 1,
            fin: false,
            writable: true,
            priority: {
              weight: 16,
              parent: 0,
              exclusive: false
            },
            path: '/',
            headers: {
              ':authority': 'localhost',
              ':path': '/',
              ':scheme': 'https',
              ':method': 'GET',

              a: 'b'
            }
          }, done);
        });
      });

      it('should generate priority request frame', function(done) {
        framer.requestFrame({
          id: 1,
          path: '/',
          host: 'localhost',
          method: 'GET',
          headers: {
            a: 'b'
          },
          priority: {
            exclusive: true,
            weight: 1
          }
        }, function(err) {
          assert(!err);

          expect({
            type: 'HEADERS',
            id: 1,
            fin: false,
            writable: true,
            priority: {
              weight: 1,
              parent: 0,

              // No exclusive flag in SPDY
              exclusive: version >= 4 ? true : false
            },
            path: '/',
            headers: {
              ':authority': 'localhost',
              ':path': '/',
              ':scheme': 'https',
              ':method': 'GET',

              a: 'b'
            }
          }, done);
        });
      });

      it('should generate fin request frame', function(done) {
        framer.requestFrame({
          id: 1,
          fin: true,
          path: '/',
          host: 'localhost',
          method: 'GET',
          headers: {
            a: 'b'
          }
        }, function(err) {
          assert(!err);

          expect({
            type: 'HEADERS',
            id: 1,
            fin: true,
            writable: true,
            priority: {
              weight: 16,
              parent: 0,
              exclusive: false
            },
            path: '/',
            headers: {
              ':authority': 'localhost',
              ':path': '/',
              ':scheme': 'https',
              ':method': 'GET',

              a: 'b'
            }
          }, done);
        });
      });

      it('should generate response frame', function(done) {
        framer.responseFrame({
          id: 1,
          status: 200,
          reason: 'OK',
          host: 'localhost',
          headers: {
            a: 'b'
          }
        }, function(err) {
          assert(!err);

          expect({
            type: 'HEADERS',
            id: 1,
            fin: false,
            writable: true,
            priority: {
              weight: 16,
              parent: 0,
              exclusive: false
            },
            path: undefined,
            headers: {
              ':status': '200',

              a: 'b'
            }
          }, done);
        });
      });

      it('should not update window on both sides', function(done) {
        framer.requestFrame({
          id: 1,
          fin: true,
          path: '/',
          host: 'localhost',
          method: 'GET',
          headers: {
            a: 'b'
          }
        }, function(err) {
          assert(!err);

          expect({
            type: 'HEADERS',
            id: 1,
            fin: true,
            writable: true,
            priority: {
              weight: 16,
              parent: 0,
              exclusive: false
            },
            path: '/',
            headers: {
              ':authority': 'localhost',
              ':path': '/',
              ':scheme': 'https',
              ':method': 'GET',

              a: 'b'
            }
          }, function() {
            assert.equal(framer.window.send.current,
                         parser.window.recv.current);
            assert.equal(framer.window.send.current, 1024 * 1024);
            done();
          });
        });
      });
    });

    describe('PUSH_PROMISE', function() {
      it('should generate regular frame', function(done) {
        framer.pushFrame({
          id: 3,
          promisedId: 41,
          path: '/',
          host: 'localhost',
          method: 'GET',
          status: 200,
          headers: {
            a: 'b'
          }
        }, function(err) {
          assert(!err);

          expect([ {
            type: 'PUSH_PROMISE',
            id: 3,
            promisedId: 41,
            fin: false,
            path: '/',
            headers: {
              ':authority': 'localhost',
              ':path': '/',
              ':scheme': 'https',
              ':method': 'GET',

              a: 'b'
            }
          }, {
            type: 'HEADERS',
            id: 41,
            priority: {
              exclusive: false,
              parent: 0,
              weight: 16
            },
            writable: true,
            path: undefined,
            fin: false,
            headers: {
              ':status': '200'
            }
          } ], done);
        });
        framer.enablePush(true);
      });

      it('should generate priority frame', function(done) {
        framer.pushFrame({
          id: 3,
          promisedId: 41,
          path: '/',
          host: 'localhost',
          method: 'GET',
          status: 200,
          priority: {
            exclusive: false,
            weight: 1,
            parent: 0
          },
          headers: {
            a: 'b'
          }
        }, function(err) {
          assert(!err);

          expect([ {
            type: 'PUSH_PROMISE',
            id: 3,
            promisedId: 41,
            fin: false,
            path: '/',
            headers: {
              ':authority': 'localhost',
              ':path': '/',
              ':scheme': 'https',
              ':method': 'GET',

              a: 'b'
            }
          }, {
            type: 'HEADERS',
            id: 41,
            priority: {
              exclusive: false,
              parent: 0,
              weight: 1
            },
            writable: true,
            path: undefined,
            fin: false,
            headers: {
              ':status': '200'
            }
          } ], done);
        });
        framer.enablePush(true);
      });

      if (version >= 4) {
        it('should fail to generate regular frame on disabled PUSH',
           function(done) {
          framer.pushFrame({
            id: 3,
            promisedId: 41,
            path: '/',
            host: 'localhost',
            method: 'GET',
            status: 200,
            headers: {
              a: 'b'
            }
          }, function(err) {
            assert(err);
            done();
          });
          framer.enablePush(false);
        });
      }
    });

    describe('trailing HEADERS', function() {
      it('should generate regular frame', function(done) {
        framer.headersFrame({
          id: 3,
          headers: {
            a: 'b'
          }
        }, function(err) {
          assert(!err);

          expect({
            type: 'HEADERS',
            id: 3,
            priority: {
              parent: 0,
              exclusive: false,
              weight: 16
            },
            fin: false,
            writable: true,
            path: undefined,
            headers: {
              a: 'b'
            }
          }, done);
        });
      });

      it('should generate frames concurrently', function(done) {
        framer.headersFrame({
          id: 3,
          headers: {
            a: 'b'
          }
        });
        framer.headersFrame({
          id: 3,
          headers: {
            c: 'd'
          }
        });

        expect([ {
          type: 'HEADERS',
          id: 3,
          priority: {
            parent: 0,
            exclusive: false,
            weight: 16
          },
          fin: false,
          writable: true,
          path: undefined,
          headers: {
            a: 'b'
          }
        }, {
          type: 'HEADERS',
          id: 3,
          priority: {
            parent: 0,
            exclusive: false,
            weight: 16
          },
          fin: false,
          writable: true,
          path: undefined,
          headers: {
            c: 'd'
          }
        } ], done);
      });

      it('should generate continuations', function(done) {
        framer.setMaxFrameSize(10);
        parser.setMaxFrameSize(10);

        framer.headersFrame({
          id: 3,
          headers: {
            a: '+++++++++++++++++++++++',
            c: '+++++++++++++++++++++++',
            e: '+++++++++++++++++++++++',
            g: '+++++++++++++++++++++++',
            i: '+++++++++++++++++++++++'
          }
        }, function(err) {
          assert(!err);

          expect({
            type: 'HEADERS',
            id: 3,
            priority: {
              parent: 0,
              exclusive: false,
              weight: 16
            },
            fin: false,
            writable: true,
            path: undefined,
            headers: {
              a: '+++++++++++++++++++++++',
              c: '+++++++++++++++++++++++',
              e: '+++++++++++++++++++++++',
              g: '+++++++++++++++++++++++',
              i: '+++++++++++++++++++++++'
            }
          }, done);
        });
      });

      it('should generate empty frame', function(done) {
        framer.headersFrame({
          id: 3,
          headers: {}
        }, function(err) {
          assert(!err);

          expect({
            type: 'HEADERS',
            id: 3,
            priority: {
              parent: 0,
              exclusive: false,
              weight: 16
            },
            fin: false,
            writable: true,
            path: undefined,
            headers: {}
          }, done);
        });
      });
    });

    describe('RST', function() {
      it('should generate regular frame', function(done) {
        framer.rstFrame({
          id: 3,
          code: 'CANCEL'
        }, function(err) {
          assert(!err);

          expect({
            type: 'RST',
            id: 3,
            code: 'CANCEL'
          }, done);
        });
      });
    });

    describe('PING', function() {
      it('should generate regular frame', function(done) {
        framer.pingFrame({
          opaque: new Buffer([ 1, 2, 3, 4, 5, 6, 7, 8 ]),
          ack: true
        }, function(err) {
          assert(!err);

          expect({
            type: 'PING',
            opaque: version < 4 ? new Buffer([ 5, 6, 7, 8 ]) :
                                  new Buffer([ 1, 2, 3, 4, 5, 6, 7, 8 ]),
            ack: true
          }, done);
        });
      });
    });

    describe('GOAWAY', function() {
      it('should generate regular frame', function(done) {
        framer.goawayFrame({
          lastId: 41,
          code: 'PROTOCOL_ERROR'
        }, function(err) {
          assert(!err);

          expect({
            type: 'GOAWAY',
            lastId: 41,
            code: 'PROTOCOL_ERROR'
          }, done);
        });
      });

      it('should generate OK frame', function(done) {
        framer.goawayFrame({
          lastId: 41,
          code: 'OK'
        }, function(err) {
          assert(!err);

          expect({
            type: 'GOAWAY',
            lastId: 41,
            code: 'OK'
          }, done);
        });
      });
    });

    describe('X_FORWARDED_FOR', function() {
      it('should generate regular frame', function(done) {
        framer.xForwardedFor({
          host: 'ok'
        }, function(err) {
          assert(!err);

          expect({
            type: 'X_FORWARDED_FOR',
            host: 'ok'
          }, done);
        });
      });
    });
  });
});
