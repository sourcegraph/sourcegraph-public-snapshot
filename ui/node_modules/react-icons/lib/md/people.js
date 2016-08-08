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

var MdPeople = function (_React$Component) {
    _inherits(MdPeople, _React$Component);

    function MdPeople() {
        _classCallCheck(this, MdPeople);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdPeople).apply(this, arguments));
    }

    _createClass(MdPeople, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm26.64 21.64q2.1883333333333326 0 4.766666666666666 0.625t4.763333333333335 2.030000000000001 2.1883333333333326 3.203333333333333v4.140000000000001h-10v-4.138333333333335q0-3.4383333333333326-3.2833333333333314-5.783333333333335 0.5500000000000007-0.07666666666666799 1.5666666666666664-0.07666666666666799z m-13.280000000000001 0q2.1883333333333326 0 4.766666666666666 0.625t4.725000000000001 2.030000000000001 2.1483333333333334 3.2049999999999983v4.140000000000001h-23.36v-4.140000000000001q-4.440892098500626e-16-1.7966666666666669 2.188333333333333-3.203333333333333t4.7666666666666675-2.0333333333333314 4.763333333333334-0.6233333333333348z m0-3.280000000000001q-2.033333333333333 0-3.5166666666666657-1.4833333333333343t-1.4833333333333343-3.5166666666666657 1.4833333333333343-3.5166666666666657 3.5166666666666657-1.4833333333333343 3.4766666666666666 1.4833333333333343 1.4450000000000003 3.5166666666666657-1.4466666666666654 3.5166666666666657-3.4766666666666666 1.4833333333333343z m13.280000000000001 0q-2.0333333333333314 0-3.5166666666666657-1.4833333333333343t-1.4833333333333343-3.5166666666666657 1.4833333333333343-3.5166666666666657 3.5166666666666657-1.4833333333333343 3.5166666666666657 1.4833333333333343 1.4833333333333343 3.5166666666666657-1.4833333333333343 3.5166666666666657-3.5166666666666657 1.4833333333333343z' })
                )
            );
        }
    }]);

    return MdPeople;
}(React.Component);

exports.default = MdPeople;
module.exports = exports['default'];