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

var TiMicrophoneOutline = function (_React$Component) {
    _inherits(TiMicrophoneOutline, _React$Component);

    function TiMicrophoneOutline() {
        _classCallCheck(this, TiMicrophoneOutline);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiMicrophoneOutline).apply(this, arguments));
    }

    _createClass(TiMicrophoneOutline, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm20 26.666666666666668c-3.676666666666666 0-6.666666666666668-2.991666666666667-6.666666666666668-6.666666666666668v-10c0-3.6750000000000007 2.9899999999999984-6.666666666666667 6.666666666666668-6.666666666666667s6.666666666666668 2.9916666666666663 6.666666666666668 6.666666666666667v10c0 3.6750000000000007-2.9899999999999984 6.666666666666668-6.666666666666668 6.666666666666668z m0-20c-1.8383333333333347 0-3.333333333333332 1.493333333333334-3.333333333333332 3.333333333333334v9.999999999999998c0 1.8399999999999999 1.495000000000001 3.333333333333332 3.333333333333332 3.333333333333332s3.333333333333332-1.4933333333333323 3.333333333333332-3.333333333333332v-10c0-1.8399999999999999-1.495000000000001-3.333333333333334-3.333333333333332-3.333333333333334z m11.666666666666668 13.333333333333332v-3.333333333333332c0-0.9216666666666669-0.745000000000001-1.666666666666666-1.6666666666666679-1.666666666666666s-1.6666666666666679 0.7449999999999992-1.6666666666666679 1.666666666666666v3.333333333333332c0 4.594999999999999-3.7383333333333333 8.333333333333336-8.333333333333332 8.333333333333336s-8.333333333333334-3.7383333333333333-8.333333333333334-8.333333333333336v-3.333333333333332c0-0.9216666666666669-0.7449999999999992-1.666666666666666-1.666666666666666-1.666666666666666s-1.666666666666666 0.7449999999999992-1.666666666666666 1.666666666666666v3.333333333333332c0 5.866666666666667 4.355 10.719999999999999 10.000000000000002 11.533333333333331v1.8000000000000043h-5.000000000000002c-0.9216666666666669 0-1.666666666666666 0.7449999999999974-1.666666666666666 1.6666666666666643s0.7449999999999992 1.6666666666666643 1.666666666666666 1.6666666666666643h13.333333333333334c0.9216666666666669 0 1.6666666666666679-0.7449999999999974 1.6666666666666679-1.6666666666666643s-0.745000000000001-1.6666666666666643-1.6666666666666679-1.6666666666666643h-5v-1.8000000000000007c5.645-0.8133333333333326 10-5.666666666666668 10-11.533333333333335z' })
                )
            );
        }
    }]);

    return TiMicrophoneOutline;
}(React.Component);

exports.default = TiMicrophoneOutline;
module.exports = exports['default'];