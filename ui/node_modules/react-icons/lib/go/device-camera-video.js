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

var GoDeviceCameraVideo = function (_React$Component) {
    _inherits(GoDeviceCameraVideo, _React$Component);

    function GoDeviceCameraVideo() {
        _classCallCheck(this, GoDeviceCameraVideo);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(GoDeviceCameraVideo).apply(this, arguments));
    }

    _createClass(GoDeviceCameraVideo, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm22.5 7.5c-1.3812500000000014 0-2.5 1.1187500000000004-2.5 2.5s1.1187499999999986 2.5 2.5 2.5 2.5-1.1187500000000004 2.5-2.5-1.1187499999999986-2.5-2.5-2.5z m12.5 7.5l-5 5v-2.5c0-1.1962499999999991-0.8399999999999999-2.196249999999999-1.9624999999999986-2.44125 1.21875-1.3337499999999984 1.9624999999999986-3.108749999999999 1.9624999999999986-5.05875 0-4.1425-3.3575000000000017-7.5-7.5-7.5-3.9499999999999993 0-7.1875 3.05375-7.47875 6.93-1.33-1.1999999999999993-3.09-1.9299999999999997-5.02125-1.9299999999999997-4.1425 0-7.5 3.3575-7.5 7.5s3.3575 7.5 7.5 7.5h-2.5v5h2.5v5c0 1.3812500000000014 1.1187500000000004 2.5 2.5 2.5h15c1.3812500000000014 0 2.5-1.1187499999999986 2.5-2.5v-2.5l5 5h2.5v-20h-2.5z m-25-2.5c-1.3812499999999996 0-2.5 1.1187500000000004-2.5 2.5s1.1187500000000004 2.5 2.5 2.5v2.5c-2.7625 0-5-2.2375000000000007-5-5s2.2375-5 5-5 5 2.2375000000000007 5 5h-2.5c0-1.3812499999999996-1.1187500000000004-2.5-2.5-2.5z m12.5 15h-5v-5h5v5z m5-4.266249999999999l-2.866250000000001-2.8674999999999997c-0.22500000000000142-0.22500000000000142-0.5375000000000014-0.36625000000000085-0.8837499999999991-0.36625000000000085h-7.5c-0.6899999999999995 0-1.25 0.5599999999999987-1.25 1.25v7.5c0 0.3249999999999993 0.125 0.6212499999999999 0.32750000000000057 0.84375l2.90625 2.90625h-4.483750000000001c-0.6912500000000001 0-1.25-0.5599999999999987-1.25-1.25v-12.5c0-0.6900000000000013 0.5587499999999999-1.25 1.25-1.25h12.5c0.6875 0 1.25 0.5599999999999987 1.25 1.25v4.483750000000001z m-5-8.23375c-2.7624999999999993 0-5-2.2375000000000007-5-5s2.2375000000000007-5 5-5 5 2.2375 5 5-2.2375000000000007 5-5 5z m12.5 12.5l-2.5-2.5 0.0037499999999965894-5.00375 2.4962500000000034-2.49625v10z' })
                )
            );
        }
    }]);

    return GoDeviceCameraVideo;
}(React.Component);

exports.default = GoDeviceCameraVideo;
module.exports = exports['default'];