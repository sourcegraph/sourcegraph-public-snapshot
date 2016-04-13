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

var MdThumbDown = function (_React$Component) {
    _inherits(MdThumbDown, _React$Component);

    function MdThumbDown() {
        _classCallCheck(this, MdThumbDown);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdThumbDown).apply(this, arguments));
    }

    _createClass(MdThumbDown, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm31.640000000000004 5h6.716666666666665v20h-6.716666666666669v-20z m-6.640000000000004 0q1.3283333333333331 0 2.3433333333333337 1.0166666666666666t1.0166666666666657 2.3416666666666677v16.641666666666666q0 1.326666666666668-1.0166666666666657 2.3416666666666686l-10.938333333333333 11.016666666666666-1.7966666666666669-1.7999999999999972q-0.7033333333333331-0.7000000000000028-0.7033333333333331-1.7166666666666686v-0.5466666666666669l1.6400000000000006-7.656666666666666h-10.545000000000002q-1.33 0-2.345-0.9766666666666666t-1.0166666666666666-2.3049999999999997l0.08000000000000007-0.1566666666666663h-0.07666666666666666v-3.203333333333333q0-0.625 0.2333333333333334-1.25l5.08-11.716666666666667q0.7833333333333332-2.033333333333333 3.046666666666667-2.033333333333333h14.999999999999998z' })
                )
            );
        }
    }]);

    return MdThumbDown;
}(React.Component);

exports.default = MdThumbDown;
module.exports = exports['default'];