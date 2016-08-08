describe('Cookie', function() {

  var cookie = require('../src/cookie.js');

  before(function() {
    cookie.reset();
  });

  afterEach(function() {
    cookie.remove('x');
    cookie.reset();
  });

  describe('get', function() {
    it('should get an existing cookie', function() {
      cookie.set('x', { a : 'b' });
      assert.deepEqual(cookie.get('x'), { a : 'b' });
    });

    it('should not throw an error on a malformed cookie', function () {
      document.cookie="x=y; path=/";
      assert.isNull(cookie.get('x'));
    });
  });

  describe('remove', function () {
    it('should remove a cookie', function() {
      cookie.set('x', { a : 'b' });
      assert.deepEqual(cookie.get('x'), { a : 'b' });
      cookie.remove('x');
      assert.isNull(cookie.get('x'));
    });
  });

  describe('options', function() {
    it('should set default options', function() {
      assert.deepEqual(cookie.options(), {
        expirationDays: undefined,
        domain: undefined
      });
    });

    it('should save options', function() {
      cookie.options({ expirationDays: 365 });
      assert.equal(cookie.options().expirationDays, 365);
    });

    it('should fallback to no domain when it cant set the test cookie', function(){
      cookie.options({ domain: 'xyz.com' });
      assert.isNull(cookie.options().domain);
      assert.isNull(cookie.get('amplitude_test'));
    });
  });
});
