'use strict';

Object.defineProperty(exports, "__esModule", {
    value: true
});

var _extends = Object.assign || function (target) { for (var i = 1; i < arguments.length; i++) { var source = arguments[i]; for (var key in source) { if (Object.prototype.hasOwnProperty.call(source, key)) { target[key] = source[key]; } } } return target; };

var _createClass = function () { function defineProperties(target, props) { for (var i = 0; i < props.length; i++) { var descriptor = props[i]; descriptor.enumerable = descriptor.enumerable || false; descriptor.configurable = true; if ("value" in descriptor) descriptor.writable = true; Object.defineProperty(target, descriptor.key, descriptor); } } return function (Constructor, protoProps, staticProps) { if (protoProps) defineProperties(Constructor.prototype, protoProps); if (staticProps) defineProperties(Constructor, staticProps); return Constructor; }; }();

function _classCallCheck(instance, Constructor) { if (!(instance instanceof Constructor)) { throw new TypeError("Cannot call a class as a function"); } }

function _possibleConstructorReturn(self, call) { if (!self) { throw new ReferenceError("this hasn't been initialised - super() hasn't been called"); } return call && (typeof call === "object" || typeof call === "function") ? call : self; }

function _inherits(subClass, superClass) { if (typeof superClass !== "function" && superClass !== null) { throw new TypeError("Super expression must either be null or a function, not " + typeof superClass); } subClass.prototype = Object.create(superClass && superClass.prototype, { constructor: { value: subClass, enumerable: false, writable: true, configurable: true } }); if (superClass) Object.setPrototypeOf ? Object.setPrototypeOf(subClass, superClass) : subClass.__proto__ = superClass; }

var React = require('react');
var IconBase = require('react-icon-base');

var GoPerson = function (_React$Component) {
    _inherits(GoPerson, _React$Component);

    function GoPerson() {
        _classCallCheck(this, GoPerson);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(GoPerson).apply(this, arguments));
    }

    _createClass(GoPerson, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm27.5 7.5c0-4.14125-3.3575000000000017-7.5-7.5-7.5s-7.5 3.3600000000000003-7.5 7.5c0 4.1425 3.3575 7.5 7.5 7.5s7.5-3.3575 7.5-7.5z m-7.5 5c-2.7600000000000016 0-5-2.24-5-5s2.240000000000002-5 5-5c2.758749999999999 0 5 2.24 5 5s-2.241250000000001 5-5 5z m5 2.5h-10c-2.76 0-5 2.240000000000002-5 5v5c0 2.758749999999999 2.24 5 5 5v10h10v-10c2.758749999999999 0 5-2.241250000000001 5-5v-5c0-2.7600000000000016-2.241250000000001-5-5-5z m2.5 10c0 1.3812500000000014-1.1187499999999986 2.5-2.5 2.5v-5h-2.5v15h-5v-15h-2.5v5c-1.3787500000000001 0-2.5-1.1187499999999986-2.5-2.5v-5c0-1.379999999999999 1.1212499999999999-2.5 2.5-2.5h10c1.3812500000000014 0 2.5 1.120000000000001 2.5 2.5v5z' })
                )
            );
        }
    }]);

    return GoPerson;
}(React.Component);

exports.default = GoPerson;
module.exports = exports['default'];