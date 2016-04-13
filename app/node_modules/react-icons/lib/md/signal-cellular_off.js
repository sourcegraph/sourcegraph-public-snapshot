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

var MdSignalCellularOff = function (_React$Component) {
    _inherits(MdSignalCellularOff, _React$Component);

    function MdSignalCellularOff() {
        _classCallCheck(this, MdSignalCellularOff);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdSignalCellularOff).apply(this, arguments));
    }

    _createClass(MdSignalCellularOff, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm7.966666666666668 7.5l28.673333333333332 28.75-2.1099999999999994 2.1099999999999994-3.2833333333333314-3.3599999999999994h-29.60666666666667l14.766666666666666-14.766666666666666-10.55-10.623333333333333z m27.03333333333333-5.859999999999999v28.671666666666667l-14.296666666666663-14.295000000000002z' })
                )
            );
        }
    }]);

    return MdSignalCellularOff;
}(React.Component);

exports.default = MdSignalCellularOff;
module.exports = exports['default'];