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

var MdOpacity = function (_React$Component) {
    _inherits(MdOpacity, _React$Component);

    function MdOpacity() {
        _classCallCheck(this, MdOpacity);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdOpacity).apply(this, arguments));
    }

    _createClass(MdOpacity, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm10 23.36h20q0-4.296666666666667-2.9666666666666686-7.266666666666666l-7.033333333333331-7.34-7.033333333333333 7.263333333333332q-2.966666666666667 2.9666666666666686-2.966666666666667 7.343333333333334z m19.453333333333337-10q3.90666666666667 3.9066666666666663 3.90666666666667 9.375 0 5.550000000000001-3.9066666666666663 9.454999999999998t-9.45333333333334 3.905000000000001-9.453333333333333-3.9066666666666663-3.9066666666666663-9.453333333333333q0-5.466666666666669 3.9066666666666663-9.373333333333333l9.453333333333333-9.455z' })
                )
            );
        }
    }]);

    return MdOpacity;
}(React.Component);

exports.default = MdOpacity;
module.exports = exports['default'];