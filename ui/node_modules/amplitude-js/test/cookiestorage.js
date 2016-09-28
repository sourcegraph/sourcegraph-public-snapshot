describe('cookieStorage', function() {
  var localStorage = require('../src/localstorage.js');
  var CookieStorage = require('../src/cookiestorage.js');
  var cookie = require('../src/cookie.js');
  var JSON = require('json'); // jshint ignore:line
  var Amplitude = require('../src/amplitude.js');
  var amplitude = new Amplitude();
  var keyPrefix = 'amp_cookiestore_';

  describe('getStorage', function() {
    it('should use cookies if enabled', function() {
      var cookieStorage = new CookieStorage();
      assert.isTrue(cookieStorage._cookiesEnabled());

      localStorage.clear();
      var uid = String(new Date());
      cookieStorage.getStorage().set(uid, uid);
      assert.equal(cookieStorage.getStorage().get(uid), uid);
      assert.equal(cookie.get(uid), uid);
      assert.equal(cookieStorage.getStorage().get(uid), cookie.get(uid));

      cookieStorage.getStorage().remove(uid);
      assert.isNull(cookieStorage.getStorage().get(uid));
      assert.isNull(cookie.get(uid));

      // assert nothing added to localstorage
      assert.isNull(localStorage.getItem(keyPrefix + uid));
    });

    it('should fall back to localstorage if cookies disabled', function() {
      var cookieStorage = new CookieStorage();
      sinon.stub(cookieStorage, '_cookiesEnabled').returns(false);
      assert.isFalse(cookieStorage._cookiesEnabled());

      localStorage.clear();
      var uid = String(new Date());
      cookieStorage.getStorage().set(uid, uid);
      assert.equal(cookieStorage.getStorage().get(uid), uid);
      assert.equal(localStorage.getItem(keyPrefix + uid), JSON.stringify(uid));
      assert.equal(cookieStorage.getStorage().get(uid), JSON.parse(localStorage.getItem(keyPrefix + uid)));

      cookieStorage.getStorage().remove(uid);
      assert.isNull(cookieStorage.getStorage().get(uid));
      assert.isNull(localStorage.getItem(keyPrefix + uid));

      // assert nothing added to cookie
      assert.isNull(cookie.get(uid));
    });

    it('should load data from localstorage if cookies disabled', function() {
      var cookieStorage = new CookieStorage();
      sinon.stub(cookieStorage, '_cookiesEnabled').returns(false);
      assert.isFalse(cookieStorage._cookiesEnabled());

      localStorage.clear();
      var uid = String(new Date());
      localStorage.setItem(keyPrefix + uid, JSON.stringify(uid));
      assert.equal(cookieStorage.getStorage().get(uid), uid)

      localStorage.removeItem(keyPrefix + uid);
      assert.isNull(cookieStorage.getStorage().get(uid));
    });
  });
});
