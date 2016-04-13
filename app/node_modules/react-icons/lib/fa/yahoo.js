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

var FaYahoo = function (_React$Component) {
    _inherits(FaYahoo, _React$Component);

    function FaYahoo() {
        _classCallCheck(this, FaYahoo);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaYahoo).apply(this, arguments));
    }

    _createClass(FaYahoo, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm22.031428571428574 21.361428571428572l0.28999999999999915 15.781428571428574q-1.3857142857142861-0.24571428571428555-2.3428571428571416-0.24571428571428555-0.9171428571428564 0-2.345714285714287 0.24571428571428555l0.28999999999999915-15.78142857142857q-0.8928571428571423-1.5399999999999991-3.7614285714285707-6.595714285714287t-4.832857142857144-8.360000000000003-4.040000000000001-6.405714285714286q1.2942857142857145 0.3342857142857143 2.41 0.3342857142857143 0.9571428571428573 0 2.475714285714286-0.3342857142857143 1.4057142857142857 2.4771428571428573 2.9800000000000004 5.122857142857143t3.7257142857142878 6.171428571428571 3.0914285714285725 5.0671428571428585q0.8257142857142838-1.361428571428572 2.442857142857143-3.9614285714285717t2.6242857142857154-4.242857142857142 2.3428571428571416-3.928571428571429 2.392857142857139-4.2285714285714295q1.2057142857142864 0.3142857142857143 2.388571428571428 0.3142857142857143 1.25 0 2.5428571428571445-0.3142857142857143-0.6242857142857119 0.8714285714285714-1.3385714285714272 1.9757142857142858t-1.1042857142857159 1.752857142857143-1.2614285714285707 2.1428571428571432-1.0928571428571416 1.8757142857142854q-3.2571428571428562 5.535714285714286-7.879999999999999 13.614285714285714z' })
                )
            );
        }
    }]);

    return FaYahoo;
}(React.Component);

exports.default = FaYahoo;
module.exports = exports['default'];