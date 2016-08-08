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

var MdSettingsBackupRestore = function (_React$Component) {
    _inherits(MdSettingsBackupRestore, _React$Component);

    function MdSettingsBackupRestore() {
        _classCallCheck(this, MdSettingsBackupRestore);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdSettingsBackupRestore).apply(this, arguments));
    }

    _createClass(MdSettingsBackupRestore, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm20 5q6.25 0 10.625 4.375t4.375 10.625-4.375 10.625-10.625 4.375q-5.156666666666666 0-9.14-3.125l2.3433333333333337-2.3433333333333337q3.1250000000000036 2.108333333333338 6.796666666666667 2.108333333333338 4.843333333333334 0 8.241666666666667-3.400000000000002t3.3999999999999986-8.240000000000002-3.3999999999999986-8.24-8.241666666666667-3.3999999999999986-8.241666666666667 3.4000000000000004-3.4000000000000004 8.239999999999998h5l-6.716666666666668 6.641666666666666-6.641666666666665-6.641666666666666h5q0-6.25 4.376666666666667-10.623333333333333t10.623333333333333-4.376666666666667z m3.3599999999999994 15q0 1.3283333333333331-1.0166666666666657 2.3433333333333337t-2.3416666666666686 1.0166666666666657-2.3416666666666686-1.0166666666666657-1.0183333333333309-2.3433333333333337 1.0166666666666657-2.3433333333333337 2.3416666666666686-1.0166666666666657 2.344999999999999 1.0166666666666657 1.0166666666666657 2.3433333333333337z' })
                )
            );
        }
    }]);

    return MdSettingsBackupRestore;
}(React.Component);

exports.default = MdSettingsBackupRestore;
module.exports = exports['default'];