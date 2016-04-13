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

var MdRemoveRedEye = function (_React$Component) {
    _inherits(MdRemoveRedEye, _React$Component);

    function MdRemoveRedEye() {
        _classCallCheck(this, MdRemoveRedEye);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdRemoveRedEye).apply(this, arguments));
    }

    _createClass(MdRemoveRedEye, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm20 15q2.0333333333333314 0 3.5166666666666657 1.4833333333333343t1.4833333333333343 3.5166666666666657-1.4833333333333343 3.5166666666666657-3.5166666666666657 1.4833333333333343-3.5166666666666657-1.4833333333333343-1.4833333333333343-3.5166666666666657 1.4833333333333343-3.5166666666666657 3.5166666666666657-1.4833333333333343z m0 13.36q3.4383333333333326 0 5.899999999999999-2.461666666666666t2.4583333333333357-5.898333333333333-2.4583333333333357-5.899999999999999-5.899999999999999-2.458333333333334-5.899999999999999 2.458333333333334-2.458333333333334 5.899999999999999 2.461666666666668 5.899999999999999 5.896666666666665 2.4583333333333357z m0-20.86q6.171666666666667 0 11.171666666666667 3.4383333333333344t7.188333333333333 9.061666666666666q-2.1883333333333326 5.626666666666665-7.188333333333333 9.06666666666667t-11.171666666666667 3.43333333333333-11.171666666666667-3.43333333333333-7.188333333333333-9.06666666666667q2.188333333333333-5.623333333333335 7.188333333333333-9.061666666666667t11.171666666666667-3.4383333333333326z' })
                )
            );
        }
    }]);

    return MdRemoveRedEye;
}(React.Component);

exports.default = MdRemoveRedEye;
module.exports = exports['default'];