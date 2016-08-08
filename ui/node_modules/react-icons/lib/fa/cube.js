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

var FaCube = function (_React$Component) {
    _inherits(FaCube, _React$Component);

    function FaCube() {
        _classCallCheck(this, FaCube);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaCube).apply(this, arguments));
    }

    _createClass(FaCube, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm21.42857142857143 36.36142857142857l14.285714285714285-7.789999999999996v-14.197142857142858l-14.285714285714285 5.199999999999999v16.785714285714285z m-1.428571428571427-19.30857142857143l15.580000000000002-5.671428571428571-15.580000000000005-5.667142857142854-15.58 5.667142857142857z m18.571428571428566-5.62428571428571v17.14285714285714q0 0.7814285714285703-0.3999999999999986 1.451428571428572t-1.095714285714287 1.048571428571428l-15.714285714285715 8.57142857142857q-0.625714285714281 0.3571428571428612-1.3614285714285685 0.3571428571428612t-1.361428571428572-0.3571428571428541l-15.714285714285715-8.57142857142857q-0.6914285714285713-0.379999999999999-1.0942857142857143-1.048571428571428t-0.4014285714285697-1.4514285714285755v-17.142857142857146q0-0.8928571428571406 0.5142857142857142-1.6285714285714263t1.3599999999999997-1.0500000000000007l15.714285714285715-5.714285714285714q0.4914285714285711-0.17857142857142838 0.9828571428571422-0.17857142857142838t0.9828571428571422 0.17857142857142838l15.714285714285715 5.714285714285714q0.8485714285714252 0.31428571428571495 1.3614285714285685 1.048571428571428t0.5128571428571433 1.6300000000000008z' })
                )
            );
        }
    }]);

    return FaCube;
}(React.Component);

exports.default = FaCube;
module.exports = exports['default'];