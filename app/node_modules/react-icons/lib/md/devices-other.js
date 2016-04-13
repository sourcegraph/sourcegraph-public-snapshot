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

var MdDevicesOther = function (_React$Component) {
    _inherits(MdDevicesOther, _React$Component);

    function MdDevicesOther() {
        _classCallCheck(this, MdDevicesOther);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdDevicesOther).apply(this, arguments));
    }

    _createClass(MdDevicesOther, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm35 30v-13.36h-6.640000000000001v13.36h6.640000000000001z m1.6400000000000006-16.64q0.625 0 1.1716666666666669 0.5083333333333329t0.5466666666666669 1.1333333333333329v16.641666666666666q0 0.625-0.5466666666666669 1.1716666666666669t-1.1716666666666669 0.5466666666666669h-10q-0.625 0-1.1333333333333329-0.5466666666666669t-0.5066666666666677-1.173333333333332v-16.641666666666666q0-0.6233333333333331 0.5066666666666677-1.1333333333333329t1.1333333333333329-0.5066666666666659h10z m-18.28 15.780000000000001q1.0166666666666657 0 1.7583333333333329-0.7416666666666671t0.7433333333333323-1.7600000000000016-0.7416666666666671-1.7583333333333329-1.7566666666666677-0.7433333333333323-1.7583333333333329 0.7416666666666671-0.7400000000000002 1.7566666666666677 0.7416666666666654 1.7583333333333329 1.7600000000000016 0.7399999999999984z m3.280000000000001-9.14v2.9666666666666686q1.7166666666666686 1.5666666666666664 1.7166666666666686 3.673333333333332 0 2.1883333333333326-1.7166666666666686 3.75v2.9666666666666686h-6.640000000000001v-2.9666666666666686q-1.6400000000000006-1.4833333333333343-1.6400000000000006-3.75 0-2.1883333333333326 1.6400000000000006-3.671666666666667v-2.9683333333333337h6.640000000000001z m-16.64-10v20h6.640000000000001v3.3599999999999994h-6.640000000000001q-1.3283333333333331 0-2.3433333333333333-1.0166666666666657t-1.0166666666666666-2.3416666666666686v-20.001666666666665q0-1.3266666666666662 1.0166666666666666-2.341666666666667t2.3433333333333333-1.0166666666666666h30v3.3616666666666672h-30z' })
                )
            );
        }
    }]);

    return MdDevicesOther;
}(React.Component);

exports.default = MdDevicesOther;
module.exports = exports['default'];