var assert = require('assert');

var transport = require('../../');
var http2 = transport.protocol.http2;

describe('HTTP2 Parser', function() {
  var parser;
  var window

  beforeEach(function() {
    window = new transport.Window({
      id: 0,
      isServer: true,
      recv: { size: 1024 * 1024 },
      send: { size: 1024 * 1024 }
    });

    var pool = http2.compressionPool.create();
    parser = http2.parser.create({
      window: window
    });
    var comp = pool.get();
    parser.setCompression(comp);
    parser.skipPreface();
  });

  function pass(data, expected, done) {
    parser.write(new Buffer(data, 'hex'));

    parser.once('data', function(frame) {
      assert.deepEqual(frame, expected);
      assert.equal(parser.buffer.size, 0);
      done();
    });
  }

  function fail(data, code, re, done) {
    parser.write(new Buffer(data, 'hex'), function(err) {
      assert(err);
      assert(err instanceof transport.protocol.base.utils.ProtocolError);
      assert.equal(err.code, http2.constants.error[code]);
      assert(re.test(err.message), err.message);

      done();
    });
  }

  describe('SETTINGS', function() {
    it('should parse regular frame', function(done) {
      pass('00000c0400000000000003000003e8000400a00000', {
        type: 'SETTINGS',
        settings: {
          initial_window_size: 10485760,
          max_concurrent_streams: 1000
        }
      }, done);
    });

    it('should parse ACK', function(done) {
      pass('000000040100000000', {
        type: 'ACK_SETTINGS'
      }, done);
    });

    it('should fail on non-empty ACK', function(done) {
      fail('000001040100000000ff', 'FRAME_SIZE_ERROR', /ACK.*non-zero/i, done);
    });

    it('should fail on non-aligned frame', function(done) {
      fail('000001040000000000ff', 'FRAME_SIZE_ERROR', /multiple/i, done);
    });

    it('should fail on non-zero stream id frame', function(done) {
      fail('000000040000000001', 'PROTOCOL_ERROR', /stream id/i, done);
    });
  });

  describe('WINDOW_UPDATE', function() {
    it('should parse regular frame', function(done) {
      pass('000004080000000000009f0001', {
        type: 'WINDOW_UPDATE',
        id: 0,
        delta: 10420225
      }, done);
    });

    it('should parse frame with negative window', function(done) {
      pass('000004080000000000ffffffff', {
        type: 'WINDOW_UPDATE',
        id: 0,
        delta: -1
      }, done);
    });

    it('should fail on bigger frame', function(done) {
      fail('000005080000000000009f000102', 'FRAME_SIZE_ERROR', /length/i, done);
    });

    it('should fail on smaller frame', function(done) {
      fail('000003080000000000009f00', 'FRAME_SIZE_ERROR', /length/i, done);
    });
  });

  describe('HEADERS', function() {
    it('should parse regular frame', function(done) {
      var hex = '000155012500000001';
      hex += '00000000ff418a089d5c0b8170dc644c8b82848753b8497ca589d34d1f43ae' +
             'ba0c41a4c7a98f33a69a3fdf9a68fa1d75d0620d263d4c79a68fbed00177fe' +
             '8d48e62b1e0b1d7f5f2c7cfdf6800bbd508e9bd9abfa5242cb40d25fa51121' +
             '27519cb2d5b6f0fab2dfbed00177be8b52dc377df6800bb9f45abefb4005da' +
             '5887a47e561cc5801f60ac8a2b5348e07dc7df10190ae171e782ebe2684b85' +
             'a0bce36cbecb8b85a642eb8f81d12e1699640f819782bbbf60bb8a2b534fb8' +
             '1f71f7c40642b85a0bce36cbecb8b8570af6a69222c83fa90d61489feffe5a' +
             '9a484aa0fea43585227fbff96a69253241fd547a8bfdff4003646e7401317a' +
             'dcd07f66a281b0dae053fad0321aa49d13fda992a49685340c8a6adca7e281' +
             '0441044cff6a435d74179163cc64b0db2eaecb8a7f59b1efd19fe94a0dd4aa' +
             '62293a9ffb52f4f61e92b0d32b817132dbab844d29b8728ec330db2eaecb9f';

      pass(hex, {
        type: 'HEADERS',
        id: 1,
        priority: {
          parent: 0,
          exclusive: false,
          weight: 256
        },
        fin: true,
        writable: true,
        path: '/',
        headers: {
          ':authority': '127.0.0.1:3232',
          ':method': 'GET',
          ':path': '/',
          ':scheme': 'https',
          accept: 'text/html,' +
                  'application/xhtml+xml,' +
                  'application/xml;q=0.9,' +
                  'image/webp,*/*;q=0.8',
          'accept-encoding': 'gzip, deflate, sdch',
          'accept-language': 'ru-RU,ru;q=0.8,en-US;q=0.6,en;q=0.4',
          'cache-control': 'max-age=0',
          cookie: '__utma=96992031.1688179242.1418653936.' +
                  '1431769072.1433090381.7; ' +
                  '__utmz=96992031.1418653936.1.1.utmcsr=(direct)|' +
                  'utmccn=(direct)|utmcmd=(none)',
          dnt: '1',
          'user-agent': 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_10_3) ' +
                        'AppleWebKit/537.36 (KHTML, like Gecko) ' +
                        'Chrome/43.0.2357.124 Safari/537.36'
        }
      }, done);
    });

    it('should unpad data', function(done) {
      var hello = '40849cb4507f84f07b2893';

      var first = '000011010c0000000105' + hello + 'ABCDEF1234';

      pass(first, {
        type: 'HEADERS',
        id: 1,
        priority: {
          parent: 0,
          exclusive: false,
          weight: 16
        },
        fin: false,
        writable: true,
        path: undefined,
        headers: {
          hello: 'world'
        }
      }, done);
    });

    it('should parse headers with continuation', function(done) {
      var hello = '40849cb4507f84f07b2893';
      var how = '40839cfe3f871d8553d1edff3f';

      var first = '00000b010000000001' + hello;
      var second = '00000d090400000001' + how;

      pass(first + second, {
        type: 'HEADERS',
        id: 1,
        priority: {
          parent: 0,
          exclusive: false,
          weight: 16
        },
        fin: false,
        writable: true,
        path: undefined,
        headers: {
          hello: 'world',
          how: 'are you?'
        }
      }, done);
    });

    it('should fail on zero stream id', function(done) {
      var first = '000000010000000000';
      fail('000000010000000000', 'PROTOCOL_ERROR', /stream id/i, done);
    });

    it('should fail to unpad too big data', function(done) {
      var hello = '40849cb4507f84f07b2893';

      var first = '000011010c00000001ff' + hello + 'ABCDEF1234';

      fail(first, 'PROTOCOL_ERROR', /invalid padding/i, done);
    });

    it('should fail to unpad too small data', function(done) {
      var first = '000000010c00000001';

      fail(first, 'FRAME_SIZE_ERROR', /not enough space/i, done);
    });

    it('should fail on just continuation', function(done) {
      var how = '40839cfe3f871d8553d1edff3f';

      var second = '00000d090400000001' + how;

      fail(second, 'PROTOCOL_ERROR', /no matching stream/i, done);
    });

    it('should fail on way double END_HEADERS continuations', function(done) {
      var hello = '40849cb4507f84f07b2893';
      var how = '40839cfe3f871d8553d1edff3f';

      var first = '00000b010000000001' + hello;
      var second = '00000d090400000001' + how;

      fail(first + second + second, 'PROTOCOL_ERROR', /no matching/i, done);
    });

    it('should fail on way too many continuations', function(done) {
      var hello = '40849cb4507f84f07b2893';
      var how = '40839cfe3f871d8553d1edff3f';

      var first = '00000b010000000001' + hello;
      var second = '00000d090000000001' + how;

      var msg = first + new Array(200).join(second);

      parser.setMaxHeaderListSize(1000);
      fail(msg, 'PROTOCOL_ERROR', /list is too large/i, done);
    });
  });

  describe('RST_STREAM', function() {
    it('should parse general frame', function(done) {
      pass('0000040300000000010000000a', {
        type: 'RST',
        id: 1,
        code: 'CONNECT_ERROR'
      }, done);
    });

    it('should fail on 0-stream', function(done) {
      fail('0000040300000000000000000a', 'PROTOCOL_ERROR', /stream id/i, done);
    });

    it('should fail on empty frame', function(done) {
      fail('000000030000000001', 'FRAME_SIZE_ERROR', /length not 4/i, done);
    });

    it('should fail on bigger frame', function(done) {
      fail('0000050300000000010102030405',
           'FRAME_SIZE_ERROR',
           /length not 4/i,
           done);
    });
  });

  describe('DATA', function() {
    it('should parse general frame', function(done) {
      pass('000004000000000001abbadead', {
        type: 'DATA',
        id: 1,
        fin: false,
        data: new Buffer('abbadead', 'hex')
      }, done);
    });

    it('should parse partial frame', function(done) {
      pass('000004000000000001abbade', {
        type: 'DATA',
        id: 1,
        fin: false,
        data: new Buffer('abbade', 'hex')
      }, function() {
        assert.equal(parser.waiting, 1);
        pass('ff', {
          type: 'DATA',
          id: 1,
          fin: false,
          data: new Buffer('ff', 'hex')
        }, function() {
          assert.equal(window.recv.current, 1048572);
          done();
        });
      });
    });

    it('should parse END_STREAM frame', function(done) {
      pass('000004000100000001deadbeef', {
        type: 'DATA',
        id: 1,
        fin: true,
        data: new Buffer('deadbeef', 'hex')
      }, done);
    });

    it('should parse partial END_STREAM frame', function(done) {
      pass('000004000100000001abbade', {
        type: 'DATA',
        id: 1,
        fin: false,
        data: new Buffer('abbade', 'hex')
      }, function() {
        assert.equal(parser.waiting, 1);
        pass('ff', {
          type: 'DATA',
          id: 1,
          fin: true,
          data: new Buffer('ff', 'hex')
        }, done);
      });
    });

    it('should parse padded frame', function(done) {
      pass('0000070008000000010212345678ffff', {
        type: 'DATA',
        id: 1,
        fin: false,
        data: new Buffer('12345678', 'hex')
      }, done);
    });

    it('should not parse partial padded frame', function(done) {
      pass('0000070008000000010212345678ff', {
        type: 'DATA',
        id: 1,
        fin: false,
        data: new Buffer('12345678', 'hex')
      }, function() {
        assert(false);
      });
      setTimeout(done, 50);
    });

    it('should fail on incorrectly padded frame', function(done) {
      fail('000007000800000001ff0000000affff',
           'PROTOCOL_ERROR',
           /padding size/i,
           done);
    });
  });

  describe('PUSH_PROMISE', function() {
    it('should parse general frame', function(done) {
      var hello = '40849cb4507f84f07b2893';

      pass('00000f05040000000100000002' + hello, {
        type: 'PUSH_PROMISE',
        id: 1,
        promisedId: 2,
        fin: false,
        headers: {
          hello: 'world'
        },
        path: undefined
      }, done);
    });

    it('should parse padded frame', function(done) {
      var hello = '40849cb4507f84f07b2893';

      pass('000011050c000000010100000002' + hello + 'ff', {
        type: 'PUSH_PROMISE',
        id: 1,
        promisedId: 2,
        fin: false,
        headers: {
          hello: 'world'
        },
        path: undefined
      }, done);
    });

    it('should fail on incorreclty padded frame', function(done) {
      var hello = '40849cb4507f84f07b2893';

      fail('000011050c00000001ff00000002' + hello + 'ff',
           'PROTOCOL_ERROR',
           /padding size/i,
           done);
    });

    it('should fail on empty frame', function(done) {
      fail('000000050400000001', 'FRAME_SIZE_ERROR', /length less than/i, done);
    });

    it('should fail on 0-stream id', function(done) {
      var hello = '40849cb4507f84f07b2893';

      fail('00000f05040000000000000002' + hello,
           'PROTOCOL_ERROR',
           /stream id/i,
           done);
    });
  });

  describe('PING', function() {
    it('should parse general frame', function(done) {
      pass('0000080600000000000102030405060708', {
        type: 'PING',
        ack: false,
        opaque: new Buffer('0102030405060708', 'hex')
      }, done);
    });

    it('should parse ack frame', function(done) {
      pass('0000080601000000000102030405060708', {
        type: 'PING',
        ack: true,
        opaque: new Buffer('0102030405060708', 'hex')
      }, done);
    });

    it('should fail on empty frame', function(done) {
      fail('000000060100000000', 'FRAME_SIZE_ERROR', /length !=/i, done);
    });

    it('should fail too big frame', function(done) {
      fail('000009060100000000010203040506070809',
           'FRAME_SIZE_ERROR',
           /length !=/i,
           done);
    });

    it('should fail on non-zero stream id', function(done) {
      fail('0000080601000000010102030405060708',
           'PROTOCOL_ERROR',
           /invalid stream id/i,
           done);
    });
  });

  describe('GOAWAY', function() {
    it('should parse general frame', function(done) {
      pass('0000080700000000000000000100000002', {
        type: 'GOAWAY',
        lastId: 1,
        code: 'INTERNAL_ERROR'
      }, done);
    });

    it('should parse frame with debug data', function(done) {
      pass('00000a0700000000000000000100000002dead', {
        type: 'GOAWAY',
        lastId: 1,
        code: 'INTERNAL_ERROR',
        debug: new Buffer('dead', 'hex')
      }, done);
    });

    it('should fail on empty frame', function(done) {
      fail('000000070000000000', 'FRAME_SIZE_ERROR', /length < 8/i, done);
    });

    it('should fail on non-zero stream id', function(done) {
      fail('0000080700000001000000000100000002',
           'PROTOCOL_ERROR',
           /invalid stream/i,
           done);
    });
  });

  describe('PRIORITY', function() {
    it('should parse general frame', function(done) {
      pass('00000502000000000100000002ab', {
        type: 'PRIORITY',
        id: 1,
        priority: {
          exclusive: false,
          parent: 2,
          weight: 0xac
        }
      }, done);
    });

    it('should parse exclusive frame', function(done) {
      pass('00000502000000000180000002ab', {
        type: 'PRIORITY',
        id: 1,
        priority: {
          exclusive: true,
          parent: 2,
          weight: 0xac
        }
      }, done);
    });

    it('should fail on empty frame', function(done) {
      fail('000000020000000001', 'FRAME_SIZE_ERROR', /length != 5/i, done);
    });

    it('should fail on too big frame', function(done) {
      fail('000006020000000001010203040506',
           'FRAME_SIZE_ERROR',
           /length != 5/i,
           done);
    });

    it('should fail on too big frame', function(done) {
      fail('000006020000000001010203040506',
           'FRAME_SIZE_ERROR',
           /length != 5/i,
           done);
    });

    it('should fail on zero stream id', function(done) {
      fail('0000050200000000000102030405',
           'PROTOCOL_ERROR',
           /invalid stream id/i,
           done);
    });
  });

  describe('X_FORWARDED_FOR', function() {
    it('should parse general frame', function(done) {
      pass('000004de00000000006f686169', {
        type: 'X_FORWARDED_FOR',
        host: 'ohai'
      }, done);
    });
  });
});
