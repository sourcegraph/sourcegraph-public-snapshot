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

var FaAngleDown = function (_React$Component) {
    _inherits(FaAngleDown, _React$Component);

    function FaAngleDown() {
        _classCallCheck(this, FaAngleDown);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaAngleDown).apply(this, arguments));
    }

    _createClass(FaAngleDown, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm31.13857142857143 16.42857142857143q0 0.28999999999999915-0.2228571428571442 0.5142857142857125l-10.399999999999999 10.399999999999999q-0.2228571428571442 0.2228571428571442-0.5142857142857125 0.2228571428571442t-0.5114285714285707-0.2228571428571442l-10.4-10.399999999999999q-0.2242857142857151-0.2242857142857133-0.2242857142857151-0.5142857142857125t0.22285714285714242-0.5142857142857142l1.1171428571428574-1.1142857142857139q0.22285714285714242-0.22285714285714242 0.5142857142857142-0.22285714285714242t0.5114285714285707 0.22285714285714242l8.768571428571425 8.77142857142857 8.774285714285718-8.77142857142857q0.2228571428571442-0.22285714285714242 0.5142857142857125-0.22285714285714242t0.5114285714285707 0.22285714285714242l1.1142857142857139 1.1142857142857139q0.2242857142857133 0.2242857142857151 0.2242857142857133 0.5142857142857142z' })
                )
            );
        }
    }]);

    return FaAngleDown;
}(React.Component);

exports.default = FaAngleDown;
module.exports = exports['default'];