'use strict';

exports.__esModule = true;
exports['default'] = stringify;

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { 'default': obj }; }

var _stringifier = require('./stringifier');

var _stringifier2 = _interopRequireDefault(_stringifier);

function stringify(node, builder) {
    var str = new _stringifier2['default'](builder);
    str.stringify(node);
}

module.exports = exports['default'];