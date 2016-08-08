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

var MdLeakAdd = function (_React$Component) {
    _inherits(MdLeakAdd, _React$Component);

    function MdLeakAdd() {
        _classCallCheck(this, MdLeakAdd);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdLeakAdd).apply(this, arguments));
    }

    _createClass(MdLeakAdd, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm23.36 35q0-4.843333333333334 3.3999999999999986-8.241666666666667t8.240000000000002-3.3999999999999986v3.2833333333333314q-3.4383333333333326 0-5.899999999999999 2.461666666666666t-2.460000000000001 5.899999999999999h-3.2833333333333314z m6.640000000000001 0q0-2.0333333333333314 1.4833333333333343-3.5166666666666657t3.5166666666666657-1.4833333333333343v5h-5z m-13.36 0q0-7.578333333333333 5.390000000000001-12.966666666666669t12.966666666666669-5.391666666666666v3.3583333333333343q-6.25 0-10.623333333333335 4.376666666666665t-4.373333333333335 10.623333333333335h-3.3633333333333333z m0-30q0 4.843333333333334-3.4000000000000004 8.241666666666667t-8.24 3.3999999999999986v-3.283333333333333q3.4383333333333344 0 5.9-2.461666666666668t2.459999999999999-5.9h3.2833333333333314z m6.719999999999999 0q0 7.578333333333333-5.390000000000001 12.966666666666669t-12.966666666666667 5.391666666666666v-3.3583333333333343q6.25 0 10.623333333333333-4.376666666666667t4.373333333333335-10.623333333333333h3.361666666666668z m-13.36 0q0 2.033333333333333-1.4833333333333343 3.5166666666666657t-3.5166666666666657 1.4833333333333343v-5h5z' })
                )
            );
        }
    }]);

    return MdLeakAdd;
}(React.Component);

exports.default = MdLeakAdd;
module.exports = exports['default'];