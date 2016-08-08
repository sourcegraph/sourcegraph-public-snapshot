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

var FaMale = function (_React$Component) {
    _inherits(FaMale, _React$Component);

    function FaMale() {
        _classCallCheck(this, FaMale);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaMale).apply(this, arguments));
    }

    _createClass(FaMale, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm31.42857142857143 15.714285714285715v9.285714285714285q0 0.8928571428571423-0.6257142857142846 1.5171428571428578t-1.5171428571428613 0.6257142857142846-1.5171428571428578-0.6257142857142846-0.6257142857142846-1.5171428571428578v-7.857142857142858h-1.428571428571427v20.357142857142858q0 1.028571428571425-0.7371428571428567 1.7628571428571433t-1.7628571428571433 0.7371428571428567-1.7628571428571433-0.7371428571428567-0.7371428571428567-1.7628571428571433v-10.357142857142858h-1.428571428571427v10.357142857142858q0 1.028571428571425-0.7371428571428567 1.7628571428571433t-1.762857142857147 0.7371428571428567-1.7628571428571433-0.7371428571428567-0.7371428571428549-1.7628571428571433v-20.357142857142858h-1.4285714285714288v7.857142857142858q0 0.8928571428571423-0.6257142857142863 1.5171428571428578t-1.517142857142856 0.6257142857142846-1.5171428571428578-0.6257142857142846-0.6257142857142863-1.5171428571428578v-9.285714285714285q0-1.7857142857142865 1.25-3.0357142857142847t3.0357142857142865-1.2500000000000018h14.285714285714288q1.7857142857142847 0 3.0357142857142847 1.25t1.25 3.0357142857142865z m-6.428571428571431-10q0 2.0757142857142856-1.46142857142857 3.5385714285714283t-3.53857142857143 1.4614285714285717-3.53857142857143-1.4614285714285717-1.46142857142857-3.538571428571429 1.46142857142857-3.5385714285714283 3.53857142857143-1.4614285714285717 3.53857142857143 1.4614285714285713 1.46142857142857 3.5385714285714287z' })
                )
            );
        }
    }]);

    return FaMale;
}(React.Component);

exports.default = FaMale;
module.exports = exports['default'];