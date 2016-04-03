'use strict';

Object.defineProperty(exports, "__esModule", {
    value: true
});

var _isObject2 = require('lodash/isObject');

var _isObject3 = _interopRequireDefault(_isObject2);

var _isFunction2 = require('lodash/isFunction');

var _isFunction3 = _interopRequireDefault(_isFunction2);

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { default: obj }; }

var ITERATOR_SYMBOL = (0, _isFunction3.default)(Symbol) && Symbol.iterator;
var OLD_ITERATOR_SYMBOL = '@@iterator';

/**
 * @see https://github.com/lodash/lodash/issues/1668
 * @see https://developer.mozilla.org/en/docs/Web/JavaScript/Reference/Iteration_protocols
 */

exports.default = function (maybeIterable) {
    var iterator = void 0;

    if (!(0, _isObject3.default)(maybeIterable)) {
        return false;
    }

    if (ITERATOR_SYMBOL) {
        iterator = maybeIterable[ITERATOR_SYMBOL];
    } else {
        iterator = maybeIterable[OLD_ITERATOR_SYMBOL];
    }

    return (0, _isFunction3.default)(iterator);
};

module.exports = exports['default'];
//# sourceMappingURL=isIterable.js.map
