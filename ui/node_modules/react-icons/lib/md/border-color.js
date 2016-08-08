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

var MdBorderColor = function (_React$Component) {
    _inherits(MdBorderColor, _React$Component);

    function MdBorderColor() {
        _classCallCheck(this, MdBorderColor);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdBorderColor).apply(this, arguments));
    }

    _createClass(MdBorderColor, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm0 33.36h40v6.640000000000001h-40v-6.640000000000001z m34.53333333333333-26.64333333333333l-3.2833333333333314 3.2833333333333314-6.25-6.25 3.2833333333333314-3.283333333333333q0.466666666666665-0.4666666666666668 1.1700000000000017-0.4666666666666668t1.1716666666666669 0.46666666666666673l3.9066666666666663 3.908333333333333q0.46666666666666856 0.4666666666666668 0.46666666666666856 1.1716666666666669t-0.46666666666666856 1.1716666666666669z m-4.923333333333332 4.923333333333332l-16.716666666666665 16.716666666666665h-6.25v-6.25l16.716666666666665-16.716666666666665z' })
                )
            );
        }
    }]);

    return MdBorderColor;
}(React.Component);

exports.default = MdBorderColor;
module.exports = exports['default'];