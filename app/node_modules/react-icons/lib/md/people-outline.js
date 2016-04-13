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

var MdPeopleOutline = function (_React$Component) {
    _inherits(MdPeopleOutline, _React$Component);

    function MdPeopleOutline() {
        _classCallCheck(this, MdPeopleOutline);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdPeopleOutline).apply(this, arguments));
    }

    _createClass(MdPeopleOutline, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm27.5 10.860000000000001q-1.3283333333333331 0-2.3433333333333337 0.9766666666666666t-1.0166666666666657 2.3049999999999997 1.0166666666666657 2.343333333333332 2.3433333333333337 1.0166666666666657 2.3433333333333337-1.0166666666666657 1.0166666666666657-2.341666666666667-1.0166666666666657-2.3049999999999997-2.3433333333333337-0.9766666666666666z m0 9.139999999999999q-2.421666666666667 0-4.140000000000001-1.7166666666666686t-1.7166666666666686-4.141666666666666 1.7166666666666686-4.100000000000001 4.140000000000001-1.6833333333333336 4.140000000000001 1.6833333333333336 1.7166666666666686 4.100000000000001-1.7166666666666686 4.140000000000001-4.140000000000001 1.7183333333333337z m-15-9.14q-1.3283333333333331 0-2.3433333333333337 0.9766666666666666t-1.0166666666666657 2.3049999999999997 1.0166666666666657 2.3433333333333337 2.3433333333333337 1.0166666666666657 2.3433333333333337-1.0166666666666657 1.0166666666666657-2.341666666666667-1.0166666666666657-2.3049999999999997-2.3433333333333337-0.9783333333333317z m0 9.14q-2.421666666666667 0-4.140000000000001-1.7166666666666686t-1.7166666666666668-4.141666666666666 1.7166666666666668-4.100000000000001 4.140000000000001-1.684999999999997 4.140000000000001 1.6833333333333336 1.7166666666666686 4.1-1.7166666666666686 4.140000000000001-4.140000000000001 1.7199999999999989z m23.36 9.14v-2.0333333333333314q0-0.7800000000000011-2.7733333333333334-1.8733333333333348t-5.586666666666666-1.0933333333333337q-2.0333333333333314 0-5 0.9383333333333326 0.8599999999999994 1.0166666666666657 0.8599999999999994 2.0333333333333314v2.030000000000001h12.5z m-15 0v-2.0333333333333314q0-0.7800000000000011-2.7733333333333334-1.8733333333333348t-5.586666666666666-1.0933333333333337-5.586666666666667 1.0933333333333337-2.7733333333333334 1.875v2.0333333333333314h16.716666666666665z m6.640000000000001-7.5q3.3599999999999994 0 7.109999999999999 1.5233333333333334t3.75 3.9450000000000003v4.533333333333335h-36.71666666666667v-4.533333333333335q2.4424906541753444e-15-2.421666666666667 3.7500000000000027-3.9450000000000003t7.108333333333333-1.5233333333333334q3.673333333333334 0 7.499999999999998 1.7166666666666686 3.8299999999999983-1.7166666666666686 7.5-1.7166666666666686z' })
                )
            );
        }
    }]);

    return MdPeopleOutline;
}(React.Component);

exports.default = MdPeopleOutline;
module.exports = exports['default'];