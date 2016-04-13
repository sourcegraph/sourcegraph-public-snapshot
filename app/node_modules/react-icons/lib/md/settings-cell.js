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

var MdSettingsCell = function (_React$Component) {
    _inherits(MdSettingsCell, _React$Component);

    function MdSettingsCell() {
        _classCallCheck(this, MdSettingsCell);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdSettingsCell).apply(this, arguments));
    }

    _createClass(MdSettingsCell, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm26.64 26.64v-20h-13.283333333333333v20h13.283333333333333z m0-26.64q1.3283333333333331 0 2.3433333333333337 1.0166666666666666t1.0166666666666657 2.34v26.643333333333334q0 1.326666666666668-1.0166666666666657 2.3416666666666686t-2.3433333333333337 1.0166666666666657h-13.283333333333333q-1.3266666666666662 0-2.341666666666667-1.0166666666666657t-1.0150000000000006-2.3416666666666686v-26.643333333333334q0-1.3283333333333323 1.0166666666666657-2.3433333333333324t2.34-1.0133333333333332h13.283333333333335z m-1.6400000000000006 40v-3.3599999999999994h3.3599999999999994v3.3599999999999994h-3.3599999999999994z m-6.640000000000001 0v-3.3599999999999994h3.2833333333333314v3.3599999999999994h-3.2833333333333314z m-6.720000000000001 0v-3.3599999999999994h3.360000000000001v3.3599999999999994h-3.3599999999999994z' })
                )
            );
        }
    }]);

    return MdSettingsCell;
}(React.Component);

exports.default = MdSettingsCell;
module.exports = exports['default'];