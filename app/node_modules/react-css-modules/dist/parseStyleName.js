'use strict';

Object.defineProperty(exports, "__esModule", {
    value: true
});

var _filter2 = require('lodash/filter');

var _filter3 = _interopRequireDefault(_filter2);

var _trim2 = require('lodash/trim');

var _trim3 = _interopRequireDefault(_trim2);

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { default: obj }; }

var styleNameIndex = {};

exports.default = function (styleNamePropertyValue, allowMultiple) {
    var styleNames = void 0;

    if (styleNameIndex[styleNamePropertyValue]) {
        styleNames = styleNameIndex[styleNamePropertyValue];
    } else {
        styleNames = (0, _trim3.default)(styleNamePropertyValue).split(' ');
        styleNames = (0, _filter3.default)(styleNames);

        styleNameIndex[styleNamePropertyValue] = styleNames;
    }

    if (allowMultiple === false && styleNames.length > 1) {
        throw new Error('ReactElement styleName property defines multiple module names ("' + styleNamePropertyValue + '").');
    }

    return styleNames;
};

module.exports = exports['default'];
//# sourceMappingURL=parseStyleName.js.map
