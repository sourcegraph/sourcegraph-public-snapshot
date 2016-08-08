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

var FaArrowLeft = function (_React$Component) {
    _inherits(FaArrowLeft, _React$Component);

    function FaArrowLeft() {
        _classCallCheck(this, FaArrowLeft);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaArrowLeft).apply(this, arguments));
    }

    _createClass(FaArrowLeft, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm36.42857142857143 20v2.8571428571428577q0 1.1828571428571415-0.7257142857142824 2.0199999999999996t-1.8857142857142861 0.8371428571428581h-15.714285714285715l6.539999999999999 6.5628571428571405q0.8485714285714288 0.8028571428571425 0.8485714285714288 2.008571428571429t-0.8485714285714288 2.008571428571429l-1.6742857142857126 1.6971428571428575q-0.8257142857142874 0.8257142857142838-2.008571428571429 0.8257142857142838-1.1600000000000001 0-2.0314285714285703-0.8257142857142838l-14.531428571428577-14.552857142857142q-0.8257142857142852-0.8242857142857147-0.8257142857142852-2.009999999999998 0-1.1571428571428584 0.8257142857142852-2.0285714285714285l14.531428571428574-14.511428571428574q0.8485714285714288-0.8485714285714279 2.0314285714285703-0.8485714285714279 1.1600000000000001 0 2.008571428571429 0.847142857142857l1.6742857142857126 1.6557142857142857q0.8485714285714288 0.847142857142857 0.8485714285714288 2.0285714285714285t-0.8485714285714288 2.032857142857143l-6.539999999999999 6.538571428571428h15.714285714285712q1.1599999999999966 0 1.8857142857142861 0.8385714285714272t0.7257142857142895 2.0185714285714305z' })
                )
            );
        }
    }]);

    return FaArrowLeft;
}(React.Component);

exports.default = FaArrowLeft;
module.exports = exports['default'];