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

var MdInput = function (_React$Component) {
    _inherits(MdInput, _React$Component);

    function MdInput() {
        _classCallCheck(this, MdInput);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdInput).apply(this, arguments));
    }

    _createClass(MdInput, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm18.36 26.64v-5h-16.716666666666665v-3.2833333333333314h16.716666666666665v-5l6.640000000000001 6.643333333333331z m16.64-21.64q1.3283333333333331 0 2.3433333333333337 0.9766666666666666t1.0166666666666657 2.383333333333333v23.36q0 1.3283333333333331-1.0166666666666657 2.3049999999999997t-2.3433333333333337 0.9750000000000014h-30q-1.3283333333333331 0-2.3433333333333333-0.9750000000000014t-1.0166666666666666-2.3049999999999997v-6.716666666666669h3.361666666666667v6.716666666666669h30v-23.436666666666667h-30.001666666666665v6.716666666666669h-3.3566666666666665v-6.639999999999999q-2.220446049250313e-16-1.3283333333333331 1.0166666666666666-2.3433333333333337t2.34-1.0166666666666675h30z' })
                )
            );
        }
    }]);

    return MdInput;
}(React.Component);

exports.default = MdInput;
module.exports = exports['default'];