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

var FaBeer = function (_React$Component) {
    _inherits(FaBeer, _React$Component);

    function FaBeer() {
        _classCallCheck(this, FaBeer);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaBeer).apply(this, arguments));
    }

    _createClass(FaBeer, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm15 20v-8.571428571428571h-5.7142857142857135v5.7142857142857135q0 1.1828571428571415 0.8371428571428563 2.0199999999999996t2.0199999999999996 0.8371428571428581h2.8571428571428577z m22.85714285714286 10v4.285714285714285h-25.71428571428572v-4.285714285714285l2.8571428571428577-4.285714285714285h-2.8571428571428577q-3.548571428571428 0-6.0600000000000005-2.5114285714285707t-2.5114285714285702-6.060000000000002v-7.142857142857142l-1.4285714285714288-1.4285714285714288 0.7142857142857144-2.8571428571428577h10.714285714285714l0.7142857142857135-2.857142857142857h21.42857142857143l0.7142857142857153 4.2857142857142865-1.4285714285714306 0.7142857142857144v17.857142857142858z' })
                )
            );
        }
    }]);

    return FaBeer;
}(React.Component);

exports.default = FaBeer;
module.exports = exports['default'];