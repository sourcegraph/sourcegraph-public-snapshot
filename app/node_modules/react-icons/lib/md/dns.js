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

var MdDns = function (_React$Component) {
    _inherits(MdDns, _React$Component);

    function MdDns() {
        _classCallCheck(this, MdDns);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdDns).apply(this, arguments));
    }

    _createClass(MdDns, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm11.64 15q1.3283333333333331 0 2.3433333333333337-1.0166666666666657t1.0166666666666657-2.3400000000000016-1.0166666666666657-2.3049999999999997-2.3433333333333337-0.9766666666666666-2.3049999999999997 0.9766666666666666-0.9749999999999996 2.3049999999999997 0.9733333333333327 2.3400000000000016 2.3083333333333336 1.0166666666666657z m21.72-10q0.7033333333333331 0 1.1716666666666669 0.4666666666666668t0.46666666666666856 1.1733333333333338v10q0 0.7033333333333331-0.46666666666666856 1.211666666666666t-1.173333333333332 0.5100000000000016h-26.715q-0.705000000000001 0-1.1733333333333338-0.5083333333333329t-0.4666666666666668-1.2100000000000009v-10q0-0.7033333333333331 0.4666666666666668-1.1716666666666669t1.171666666666666-0.47166666666666757h26.71666666666667z m-21.72 26.640000000000004q1.3283333333333331 0 2.3433333333333337-0.9766666666666666t1.0166666666666657-2.3050000000000033-1.0166666666666657-2.3433333333333337-2.3433333333333337-1.0150000000000006-2.3049999999999997 1.0166666666666657-0.9749999999999996 2.3416666666666686 0.9766666666666666 2.3049999999999997 2.3066666666666666 0.9766666666666666z m21.72-10q0.7033333333333331 0 1.1716666666666669 0.5083333333333329t0.46666666666666856 1.2100000000000009v9.999999999999996q0 0.7033333333333331-0.46666666666666856 1.1716666666666669t-1.173333333333332 0.46666666666666856h-26.715q-0.705000000000001 0-1.1733333333333338-0.46666666666666856t-0.4666666666666668-1.173333333333332v-10q0-0.7033333333333331 0.4666666666666668-1.211666666666666t1.1716666666666669-0.5100000000000016h26.71666666666667z' })
                )
            );
        }
    }]);

    return MdDns;
}(React.Component);

exports.default = MdDns;
module.exports = exports['default'];