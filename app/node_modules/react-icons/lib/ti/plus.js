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

var TiPlus = function (_React$Component) {
    _inherits(TiPlus, _React$Component);

    function TiPlus() {
        _classCallCheck(this, TiPlus);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiPlus).apply(this, arguments));
    }

    _createClass(TiPlus, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm30 16.666666666666668h-6.666666666666668v-6.666666666666668c0-1.8399999999999999-1.4933333333333323-3.333333333333334-3.333333333333332-3.333333333333334s-3.333333333333332 1.493333333333334-3.333333333333332 3.333333333333334l0.11666666666666714 6.666666666666668h-6.783333333333335c-1.8399999999999999 0-3.333333333333334 1.4933333333333323-3.333333333333334 3.333333333333332s1.493333333333334 3.333333333333332 3.333333333333334 3.333333333333332l6.783333333333335-0.11666666666666714-0.11666666666666714 6.783333333333335c0 1.8399999999999999 1.4933333333333323 3.3333333333333357 3.333333333333332 3.3333333333333357s3.333333333333332-1.4933333333333323 3.333333333333332-3.333333333333332v-6.783333333333335l6.666666666666668 0.11666666666666714c1.8399999999999999 0 3.3333333333333357-1.4933333333333323 3.3333333333333357-3.333333333333332s-1.4933333333333323-3.333333333333332-3.333333333333332-3.333333333333332z' })
                )
            );
        }
    }]);

    return TiPlus;
}(React.Component);

exports.default = TiPlus;
module.exports = exports['default'];