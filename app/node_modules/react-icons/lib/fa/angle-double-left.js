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

var FaAngleDoubleLeft = function (_React$Component) {
    _inherits(FaAngleDoubleLeft, _React$Component);

    function FaAngleDoubleLeft() {
        _classCallCheck(this, FaAngleDoubleLeft);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaAngleDoubleLeft).apply(this, arguments));
    }

    _createClass(FaAngleDoubleLeft, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm21.852857142857143 30.714285714285715q0 0.28999999999999915-0.2228571428571442 0.5142857142857125l-1.1142857142857139 1.1142857142857139q-0.2228571428571442 0.2228571428571442-0.514285714285716 0.2228571428571442t-0.5114285714285707-0.2228571428571442l-10.4-10.399999999999999q-0.2271428571428551-0.22571428571428598-0.2271428571428551-0.5142857142857125t0.22285714285714242-0.5142857142857125l10.402857142857144-10.4q0.2228571428571442-0.22285714285714242 0.514285714285716-0.22285714285714242t0.5114285714285707 0.22285714285714242l1.1142857142857139 1.1142857142857139q0.2242857142857133 0.2242857142857151 0.2242857142857133 0.5142857142857142t-0.2228571428571442 0.5142857142857142l-8.772857142857141 8.77142857142857 8.771428571428569 8.771428571428572q0.2242857142857133 0.2242857142857133 0.2242857142857133 0.5142857142857125z m8.571428571428573 0q0 0.28999999999999915-0.2228571428571442 0.5142857142857125l-1.1142857142857139 1.1142857142857139q-0.2228571428571442 0.2228571428571442-0.5142857142857125 0.2228571428571442t-0.5114285714285707-0.2228571428571442l-10.400000000000002-10.399999999999999q-0.2242857142857133-0.2242857142857133-0.2242857142857133-0.5142857142857125t0.2228571428571442-0.5142857142857125l10.402857142857144-10.4q0.2228571428571442-0.22285714285714242 0.5142857142857125-0.22285714285714242t0.5114285714285707 0.22285714285714242l1.1142857142857139 1.1142857142857139q0.2242857142857133 0.2242857142857151 0.2242857142857133 0.5142857142857142t-0.2228571428571442 0.5142857142857142l-8.775714285714283 8.77142857142857 8.771428571428572 8.771428571428572q0.2242857142857133 0.2242857142857133 0.2242857142857133 0.5142857142857125z' })
                )
            );
        }
    }]);

    return FaAngleDoubleLeft;
}(React.Component);

exports.default = FaAngleDoubleLeft;
module.exports = exports['default'];