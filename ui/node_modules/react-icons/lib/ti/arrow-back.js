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

var TiArrowBack = function (_React$Component) {
    _inherits(TiArrowBack, _React$Component);

    function TiArrowBack() {
        _classCallCheck(this, TiArrowBack);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiArrowBack).apply(this, arguments));
    }

    _createClass(TiArrowBack, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm20 15.100000000000001v-4.2666666666666675c0-0.42666666666666586-0.163333333333334-0.8533333333333335-0.4883333333333333-1.1799999999999997-0.3249999999999993-0.3249999999999993-0.75-0.48666666666666636-1.1783333333333346-0.48666666666666636s-0.8533333333333317 0.16166666666666707-1.1783333333333346 0.48666666666666636l-10.48833333333333 10.346666666666666 10.488333333333333 10.344999999999999c0.3249999999999993 0.3249999999999993 0.75 0.4883333333333333 1.1783333333333346 0.4883333333333333s0.8533333333333317-0.163333333333334 1.1783333333333346-0.4883333333333333 0.4883333333333333-0.7533333333333339 0.4883333333333333-1.1783333333333346v-4.149999999999999c4.583333333333332 0.11666666666666714 9.591666666666665 0.9450000000000003 13.333333333333332 6.649999999999999v-1.6666666666666679c0-7.721666666666668-5.833333333333336-14.071666666666667-13.333333333333336-14.9z' })
                )
            );
        }
    }]);

    return TiArrowBack;
}(React.Component);

exports.default = TiArrowBack;
module.exports = exports['default'];