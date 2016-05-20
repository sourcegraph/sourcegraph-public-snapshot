'use strict';

Object.defineProperty(exports, "__esModule", {
  value: true
});

var _typeof = typeof Symbol === "function" && typeof Symbol.iterator === "symbol" ? function (obj) { return typeof obj; } : function (obj) { return obj && typeof Symbol === "function" && obj.constructor === Symbol ? "symbol" : typeof obj; };

exports.default = createAnimateManager;

var _setRafTimeout = require('./setRafTimeout');

var _setRafTimeout2 = _interopRequireDefault(_setRafTimeout);

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { default: obj }; }

function _toArray(arr) { return Array.isArray(arr) ? arr : Array.from(arr); }

function createAnimateManager() {
  var currStyle = {};
  var handleChange = function handleChange() {
    return null;
  };
  var shouldStop = false;

  var setStyle = function setStyle(_style) {
    if (shouldStop) {
      return;
    }

    if (Array.isArray(_style)) {
      if (!_style.length) {
        return;
      }

      var styles = _style;

      var _styles = _toArray(styles);

      var curr = _styles[0];

      var restStyles = _styles.slice(1);

      if (typeof curr === 'number') {
        (0, _setRafTimeout2.default)(setStyle.bind(null, restStyles), curr);

        return;
      }

      setStyle(curr);
      (0, _setRafTimeout2.default)(setStyle.bind(null, restStyles));
      return;
    }

    if ((typeof _style === 'undefined' ? 'undefined' : _typeof(_style)) === 'object') {
      currStyle = _style;
      handleChange(currStyle);
    }

    if (typeof _style === 'function') {
      _style();
    }
  };

  return {
    stop: function stop() {
      shouldStop = true;
    },
    start: function start(style) {
      shouldStop = false;
      setStyle(style);
    },
    subscribe: function subscribe(_handleChange) {
      handleChange = _handleChange;

      return function () {
        handleChange = function handleChange() {
          return null;
        };
      };
    }
  };
}