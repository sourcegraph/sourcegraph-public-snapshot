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

var MdPowerSettingsNew = function (_React$Component) {
    _inherits(MdPowerSettingsNew, _React$Component);

    function MdPowerSettingsNew() {
        _classCallCheck(this, MdPowerSettingsNew);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdPowerSettingsNew).apply(this, arguments));
    }

    _createClass(MdPowerSettingsNew, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm29.688333333333333 8.593333333333334q5.311666666666667 4.533333333333333 5.311666666666667 11.406666666666666 0 6.25-4.373333333333335 10.625t-10.626666666666665 4.375-10.623333333333333-4.375-4.376666666666667-10.625q0-6.875 5.316666666666666-11.406666666666668l2.341666666666667 2.3433333333333337q-4.296666666666665 3.5166666666666693-4.296666666666665 9.063333333333334 0 4.843333333333334 3.4000000000000004 8.241666666666667t8.238333333333332 3.3999999999999986 8.243333333333332-3.3999999999999986 3.400000000000002-8.241666666666667q0-5.546666666666667-4.300000000000001-8.983333333333333z m-8.049999999999997-3.5933333333333337v16.64h-3.280000000000001v-16.64h3.2833333333333314z' })
                )
            );
        }
    }]);

    return MdPowerSettingsNew;
}(React.Component);

exports.default = MdPowerSettingsNew;
module.exports = exports['default'];