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

var FaArrowUp = function (_React$Component) {
    _inherits(FaArrowUp, _React$Component);

    function FaArrowUp() {
        _classCallCheck(this, FaArrowUp);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaArrowUp).apply(this, arguments));
    }

    _createClass(FaArrowUp, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm37.38857142857143 21.674285714285716q0 1.138571428571428-0.8257142857142838 2.008571428571429l-1.6742857142857162 1.6742857142857126q-0.8485714285714252 0.8485714285714288-2.0314285714285703 0.8485714285714288-1.2057142857142864 0-2.008571428571429-0.8485714285714288l-6.562857142857148-6.539999999999999v15.714285714285719q0 1.1599999999999966-0.8371428571428581 1.8857142857142861t-2.019999999999996 0.7257142857142824h-2.8571428571428577q-1.1828571428571415 0-2.0199999999999996-0.7257142857142824t-0.8371428571428581-1.8857142857142932v-15.714285714285715l-6.562857142857144 6.540000000000003q-0.8028571428571425 0.8485714285714288-2.008571428571429 0.8485714285714288t-2.008571428571429-0.8485714285714288l-1.6742857142857144-1.6742857142857126q-0.8485714285714288-0.8485714285714288-0.8485714285714288-2.008571428571429 0-1.1828571428571415 0.8485714285714288-2.0314285714285703l14.531428571428572-14.531428571428574q0.7800000000000011-0.8257142857142865 2.008571428571429-0.8257142857142865 1.2057142857142864 0 2.0314285714285703 0.8257142857142856l14.531428571428577 14.53142857142857q0.8257142857142838 0.8714285714285701 0.8257142857142838 2.0314285714285703z' })
                )
            );
        }
    }]);

    return FaArrowUp;
}(React.Component);

exports.default = FaArrowUp;
module.exports = exports['default'];