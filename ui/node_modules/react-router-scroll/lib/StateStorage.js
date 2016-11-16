'use strict';

exports.__esModule = true;

var _DOMStateStorage = require('history/lib/DOMStateStorage');

function _classCallCheck(instance, Constructor) { if (!(instance instanceof Constructor)) { throw new TypeError("Cannot call a class as a function"); } }

var STATE_KEY_PREFIX = '@@scroll|';

var StateStorage = function () {
  function StateStorage(router) {
    _classCallCheck(this, StateStorage);

    this.getFallbackLocationKey = router.createPath;
  }

  StateStorage.prototype.read = function read(location, key) {
    return (0, _DOMStateStorage.readState)(this.getStateKey(location, key));
  };

  StateStorage.prototype.save = function save(location, key, value) {
    (0, _DOMStateStorage.saveState)(this.getStateKey(location, key), value);
  };

  StateStorage.prototype.getStateKey = function getStateKey(location, key) {
    var locationKey = location.key || this.getFallbackLocationKey(location);
    var stateKeyBase = '' + STATE_KEY_PREFIX + locationKey;
    return key == null ? stateKeyBase : stateKeyBase + '|' + key;
  };

  return StateStorage;
}();

exports.default = StateStorage;
module.exports = exports['default'];