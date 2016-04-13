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

var MdPhoneAndroid = function (_React$Component) {
    _inherits(MdPhoneAndroid, _React$Component);

    function MdPhoneAndroid() {
        _classCallCheck(this, MdPhoneAndroid);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdPhoneAndroid).apply(this, arguments));
    }

    _createClass(MdPhoneAndroid, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm28.75 30v-23.36h-17.5v23.36h17.5z m-5.390000000000001 5v-1.6400000000000006h-6.716666666666669v1.6400000000000006h6.716666666666669z m3.280000000000001-33.36q2.0333333333333314-4.440892098500626e-16 3.5166666666666657 1.483333333333333t1.4833333333333343 3.516666666666667v26.71666666666667q0 2.0333333333333314-1.4833333333333343 3.5166666666666657t-3.5166666666666657 1.4833333333333343h-13.283333333333333q-2.0299999999999994 0-3.5133333333333336-1.4833333333333343t-1.4833333333333343-3.5166666666666657v-26.715000000000003q0-2.0333333333333323 1.4833333333333343-3.5166666666666657t3.5166666666666657-1.4833333333333334h13.280000000000001z' })
                )
            );
        }
    }]);

    return MdPhoneAndroid;
}(React.Component);

exports.default = MdPhoneAndroid;
module.exports = exports['default'];