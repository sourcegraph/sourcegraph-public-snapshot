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

var FaThLarge = function (_React$Component) {
    _inherits(FaThLarge, _React$Component);

    function FaThLarge() {
        _classCallCheck(this, FaThLarge);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaThLarge).apply(this, arguments));
    }

    _createClass(FaThLarge, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm18.571428571428573 22.857142857142858v8.571428571428573q0 1.1599999999999966-0.8485714285714288 2.008571428571429t-2.008571428571429 0.8485714285714252h-11.42857142857143q-1.1599999999999993 0-2.0085714285714276-0.8485714285714252t-0.8485714285714285-2.008571428571429v-8.57142857142857q0-1.1600000000000001 0.8485714285714285-2.008571428571429t2.0085714285714285-0.8485714285714323h11.42857142857143q1.1600000000000001 0 2.008571428571429 0.8485714285714288t0.8485714285714288 2.008571428571429z m0-17.142857142857142v8.571428571428571q0 1.1600000000000001-0.8485714285714288 2.008571428571427t-2.008571428571429 0.8485714285714288h-11.42857142857143q-1.1599999999999993 0-2.0085714285714276-0.8485714285714288t-0.8485714285714285-2.008571428571427v-8.571428571428573q0-1.1599999999999993 0.8485714285714285-2.0085714285714276t2.0085714285714285-0.8485714285714288h11.42857142857143q1.1600000000000001 0 2.008571428571429 0.8485714285714283t0.8485714285714288 2.008571428571429z m19.999999999999996 17.142857142857142v8.571428571428573q0 1.1599999999999966-0.8485714285714252 2.008571428571429t-2.008571428571429 0.8485714285714252h-11.42857142857143q-1.1600000000000001 0-2.008571428571429-0.8485714285714252t-0.8485714285714252-2.008571428571429v-8.57142857142857q0-1.1600000000000001 0.8485714285714288-2.008571428571429t2.0085714285714253-0.8485714285714323h11.42857142857143q1.1600000000000037 0 2.008571428571429 0.8485714285714288t0.8485714285714252 2.008571428571429z m0-17.142857142857142v8.571428571428571q0 1.1600000000000001-0.8485714285714252 2.008571428571427t-2.008571428571429 0.8485714285714288h-11.42857142857143q-1.1600000000000001 0-2.008571428571429-0.8485714285714288t-0.8485714285714252-2.008571428571427v-8.571428571428573q0-1.1599999999999993 0.8485714285714288-2.0085714285714276t2.0085714285714253-0.8485714285714288h11.42857142857143q1.1600000000000037 0 2.008571428571429 0.8485714285714283t0.8485714285714252 2.008571428571429z' })
                )
            );
        }
    }]);

    return FaThLarge;
}(React.Component);

exports.default = FaThLarge;
module.exports = exports['default'];