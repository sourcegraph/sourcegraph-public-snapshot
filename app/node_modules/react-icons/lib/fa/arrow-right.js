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

var FaArrowRight = function (_React$Component) {
    _inherits(FaArrowRight, _React$Component);

    function FaArrowRight() {
        _classCallCheck(this, FaArrowRight);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaArrowRight).apply(this, arguments));
    }

    _createClass(FaArrowRight, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm36.42857142857143 21.42857142857143q0 1.2057142857142864-0.8257142857142838 2.0314285714285703l-14.531428571428574 14.53142857142857q-0.8714285714285701 0.8257142857142838-2.0314285714285703 0.8257142857142838-1.138571428571428 0-2.008571428571429-0.8257142857142838l-1.6742857142857162-1.674285714285709q-0.8485714285714288-0.8485714285714252-0.8485714285714288-2.0314285714285703t0.8485714285714288-2.0314285714285703l6.539999999999999-6.540000000000006h-15.714285714285715q-1.1600000000000001 0-1.8857142857142861-0.8371428571428581t-0.7257142857142838-2.0199999999999996v-2.8571428571428577q0-1.1828571428571415 0.7257142857142855-2.0199999999999996t1.8857142857142861-0.8371428571428581h15.714285714285717l-6.540000000000003-6.562857142857142q-0.8485714285714305-0.8028571428571425-0.8485714285714305-2.008571428571429t0.8485714285714288-2.008571428571429l1.6742857142857144-1.6742857142857144q0.8485714285714288-0.8485714285714288 2.008571428571429-0.8485714285714288 1.1828571428571415 0 2.0314285714285703 0.8485714285714288l14.53142857142857 14.531428571428574q0.825714285714291 0.7800000000000011 0.825714285714291 2.008571428571429z' })
                )
            );
        }
    }]);

    return FaArrowRight;
}(React.Component);

exports.default = FaArrowRight;
module.exports = exports['default'];