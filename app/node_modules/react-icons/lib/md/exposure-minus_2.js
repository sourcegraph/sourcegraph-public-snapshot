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

var MdExposureMinus2 = function (_React$Component) {
    _inherits(MdExposureMinus2, _React$Component);

    function MdExposureMinus2() {
        _classCallCheck(this, MdExposureMinus2);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdExposureMinus2).apply(this, arguments));
    }

    _createClass(MdExposureMinus2, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm3.3600000000000003 18.36h13.283333333333335v3.2833333333333314h-13.283333333333333v-3.2833333333333314z m24.061666666666667-10q6.716666666666669 0 6.716666666666669 5.859999999999999 0 1.0933333333333337-0.3116666666666674 1.9533333333333331-0.5466666666666669 1.4833333333333343-0.8599999999999994 1.9533333333333331-1.3283333333333331 2.1099999999999994-3.125 3.9066666666666663l-4.766666666666666 5.156666666666666h9.924999999999997v2.8133333333333326h-14.376666666666665v-2.5l6.953333333333333-7.578333333333333q1.4066666666666663-1.4083333333333314 2.423333333333332-3.1249999999999964 0.5450000000000017-0.9383333333333326 0.5450000000000017-2.1883333333333326 0-1.0166666666666657-0.1566666666666663-1.4066666666666663-0.783333333333335-2.033333333333333-2.9666666666666686-2.033333333333333-3.594999999999999 0-3.594999999999999 3.83h-3.5933333333333337q0-2.8116666666666674 1.875-4.686666666666667 1.9533333333333331-1.9533333333333331 5.313333333333333-1.9533333333333331z' })
                )
            );
        }
    }]);

    return MdExposureMinus2;
}(React.Component);

exports.default = MdExposureMinus2;
module.exports = exports['default'];