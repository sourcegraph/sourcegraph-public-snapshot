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

var FaChild = function (_React$Component) {
    _inherits(FaChild, _React$Component);

    function FaChild() {
        _classCallCheck(this, FaChild);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaChild).apply(this, arguments));
    }

    _createClass(FaChild, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm32.23142857142857 12.232857142857144l-6.517142857142858 6.517142857142856v18.392857142857146q0 1.028571428571432-0.7371428571428567 1.7628571428571433t-1.7628571428571433 0.7371428571428567-1.7628571428571433-0.7371428571428567-0.7371428571428567-1.7628571428571433v-8.57142857142857h-1.428571428571427v8.57142857142857q0 1.028571428571432-0.7371428571428567 1.7628571428571433t-1.7628571428571433 0.7371428571428567-1.7628571428571433-0.7371428571428567-0.7371428571428584-1.7628571428571433v-18.392857142857146l-6.517142857142858-6.517142857142856q-0.6257142857142854-0.6257142857142863-0.6257142857142854-1.5185714285714287t0.6257142857142854-1.5142857142857142 1.5171428571428578-0.6285714285714299 1.5171428571428578 0.6285714285714281l5.089999999999998 5.085714285714287h8.214285714285715l5.09-5.090000000000002q0.6257142857142881-0.6242857142857137 1.5171428571428578-0.6242857142857137t1.5171428571428578 0.6285714285714281 0.6257142857142881 1.514285714285716-0.6257142857142881 1.5185714285714287z m-7.231428571428573-3.661428571428573q0 2.0757142857142856-1.46142857142857 3.5385714285714283t-3.53857142857143 1.4614285714285717-3.53857142857143-1.46142857142857-1.46142857142857-3.53857142857143 1.46142857142857-3.5385714285714283 3.53857142857143-1.4614285714285713 3.53857142857143 1.4614285714285713 1.46142857142857 3.5385714285714283z' })
                )
            );
        }
    }]);

    return FaChild;
}(React.Component);

exports.default = FaChild;
module.exports = exports['default'];