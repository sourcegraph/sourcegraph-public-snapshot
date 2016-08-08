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

var MdAlarmOn = function (_React$Component) {
    _inherits(MdAlarmOn, _React$Component);

    function MdAlarmOn() {
        _classCallCheck(this, MdAlarmOn);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdAlarmOn).apply(this, arguments));
    }

    _createClass(MdAlarmOn, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm17.578333333333337 24.216666666666665l8.203333333333333-8.280000000000001 1.7949999999999982 1.7966666666666669-10 10-5.313333333333333-5.313333333333333 1.7166666666666668-1.716666666666665z m2.4216666666666633 9.145q4.843333333333334 0 8.241666666666667-3.4383333333333326t3.3999999999999986-8.283333333333331-3.3999999999999986-8.24-8.241666666666667-3.4000000000000004-8.241666666666667 3.3999999999999986-3.4000000000000004 8.240000000000002 3.4000000000000004 8.283333333333331 8.241666666666667 3.4350000000000023z m0-26.720000000000002q6.25 0 10.625 4.413333333333334t4.375 10.585000000000004-4.375 10.586666666666666-10.625 4.413333333333341-10.625-4.413333333333334-4.375-10.586666666666673 4.375-10.586666666666668 10.625-4.413333333333332z m-6.875-1.0166666666666666l-7.658333333333333 6.408333333333338-2.1083333333333334-2.5 7.658333333333334-6.41z m23.516666666666666 3.908333333333334l-2.1116666666666646 2.576666666666666-7.656666666666666-6.483333333333333 2.1099999999999994-2.5z' })
                )
            );
        }
    }]);

    return MdAlarmOn;
}(React.Component);

exports.default = MdAlarmOn;
module.exports = exports['default'];