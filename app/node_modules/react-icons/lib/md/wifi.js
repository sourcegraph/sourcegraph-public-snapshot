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

var MdWifi = function (_React$Component) {
    _inherits(MdWifi, _React$Component);

    function MdWifi() {
        _classCallCheck(this, MdWifi);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdWifi).apply(this, arguments));
    }

    _createClass(MdWifi, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm8.360000000000001 21.64q4.843333333333334-4.766666666666666 11.679999999999998-4.766666666666666t11.600000000000001 4.766666666666666l-3.2783333333333324 3.3599999999999994q-3.4400000000000013-3.4383333333333326-8.361666666666668-3.4383333333333326t-8.36 3.4383333333333326z m6.639999999999999 6.719999999999999q2.0333333333333314-2.0333333333333314 5-2.0333333333333314t5 2.0333333333333314l-5 5z m-13.36-13.36q7.656666666666666-7.578333333333333 18.4-7.578333333333333t18.315000000000005 7.578333333333333l-3.355000000000004 3.3599999999999994q-6.25-6.171666666666667-15-6.171666666666667t-15 6.171666666666667z' })
                )
            );
        }
    }]);

    return MdWifi;
}(React.Component);

exports.default = MdWifi;
module.exports = exports['default'];