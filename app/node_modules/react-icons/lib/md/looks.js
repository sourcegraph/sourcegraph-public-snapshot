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

var MdLooks = function (_React$Component) {
    _inherits(MdLooks, _React$Component);

    function MdLooks() {
        _classCallCheck(this, MdLooks);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdLooks).apply(this, arguments));
    }

    _createClass(MdLooks, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm20 10q7.578333333333333 0 12.966666666666669 5.390000000000001t5.391666666666666 12.966666666666669h-3.3583333333333343q0-6.170000000000002-4.376666666666665-10.583333333333336t-10.623333333333335-4.418333333333333-10.626666666666667 4.416666666666668-4.373333333333333 10.583333333333336h-3.3633333333333333q0-7.576666666666668 5.390000000000001-12.966666666666667t12.968333333333334-5.388333333333337z m0 6.640000000000001q4.766666666666666 0 8.203333333333333 3.4383333333333326t3.4383333333333326 8.283333333333331h-3.2833333333333314q0-3.4400000000000013-2.460000000000001-5.899999999999999t-5.898333333333333-2.461666666666666-5.9 2.461666666666666-2.459999999999999 5.899999999999999h-3.283333333333333q0-4.844999999999999 3.4399999999999995-8.283333333333331t8.203333333333333-3.4383333333333326z' })
                )
            );
        }
    }]);

    return MdLooks;
}(React.Component);

exports.default = MdLooks;
module.exports = exports['default'];