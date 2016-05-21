'use strict';

Object.defineProperty(exports, "__esModule", {
  value: true
});
exports.default = setRafTimeout;

var _raf = require('raf');

var _raf2 = _interopRequireDefault(_raf);

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { default: obj }; }

function setRafTimeout(callback) {
  var timeout = arguments.length <= 1 || arguments[1] === undefined ? 0 : arguments[1];

  var currTime = -1;

  var shouldUpdate = function shouldUpdate(now) {
    if (currTime < 0) {
      currTime = now;
    }

    if (now - currTime > timeout) {
      callback(now);
      currTime = -1;
    } else {
      (0, _raf2.default)(shouldUpdate);
    }
  };

  (0, _raf2.default)(shouldUpdate);
}