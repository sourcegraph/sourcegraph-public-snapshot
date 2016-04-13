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

var FaSitemap = function (_React$Component) {
    _inherits(FaSitemap, _React$Component);

    function FaSitemap() {
        _classCallCheck(this, FaSitemap);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaSitemap).apply(this, arguments));
    }

    _createClass(FaSitemap, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm40 27.857142857142858v7.142857142857142q0 0.8928571428571459-0.6257142857142881 1.5171428571428578t-1.5171428571428507 0.6257142857142881h-7.142857142857142q-0.8928571428571423 0-1.5171428571428578-0.6257142857142881t-0.6257142857142881-1.5171428571428578v-7.142857142857142q0-0.8928571428571423 0.6257142857142846-1.5171428571428578t1.5171428571428578-0.6257142857142846h2.142857142857146v-4.285714285714285h-11.42857142857143v4.285714285714285h2.1428571428571423q0.8928571428571423 0 1.5171428571428578 0.6257142857142846t0.6257142857142846 1.5171428571428578v7.142857142857142q0 0.8928571428571459-0.6257142857142846 1.5171428571428578t-1.5171428571428578 0.6257142857142881h-7.142857142857142q-0.8928571428571423 0-1.5171428571428578-0.6257142857142881t-0.6257142857142863-1.5171428571428578v-7.142857142857142q0-0.8928571428571423 0.6257142857142863-1.5171428571428578t1.5171428571428578-0.6257142857142846h2.1428571428571423v-4.285714285714285h-11.42857142857143v4.285714285714285h2.142857142857144q0.8928571428571423 0 1.5171428571428578 0.6257142857142846t0.6257142857142846 1.5171428571428578v7.142857142857142q0 0.8928571428571459-0.6257142857142863 1.5171428571428578t-1.517142857142856 0.6257142857142881h-7.142857142857143q-0.8928571428571428 0-1.5171428571428573-0.6257142857142881t-0.6257142857142859-1.5171428571428578v-7.142857142857142q0-0.8928571428571423 0.6257142857142858-1.5171428571428578t1.517142857142857-0.6257142857142846h2.142857142857143v-4.285714285714285q0-1.1600000000000001 0.8485714285714288-2.008571428571429t2.008571428571429-0.8485714285714288h11.42857142857143v-4.285714285714285h-2.1428571428571423q-0.8928571428571423 0-1.5171428571428578-0.6257142857142863t-0.6257142857142863-1.5171428571428596v-7.142857142857143q0-0.8928571428571432 0.6257142857142863-1.5171428571428573t1.5171428571428578-0.6257142857142846h7.142857142857142q0.8928571428571423 0 1.5171428571428578 0.6257142857142859t0.6257142857142846 1.517142857142857v7.142857142857142q0 0.8928571428571423-0.6257142857142846 1.5171428571428578t-1.5171428571428578 0.6257142857142863h-2.1428571428571423v4.2857142857142865h11.42857142857143q1.1599999999999966 0 2.008571428571429 0.8485714285714288t0.8485714285714252 2.008571428571429v4.285714285714285h2.142857142857146q0.8928571428571459 0 1.5171428571428578 0.6257142857142846t0.625714285714281 1.5171428571428578z' })
                )
            );
        }
    }]);

    return FaSitemap;
}(React.Component);

exports.default = FaSitemap;
module.exports = exports['default'];