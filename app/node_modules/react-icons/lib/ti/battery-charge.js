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

var TiBatteryCharge = function (_React$Component) {
    _inherits(TiBatteryCharge, _React$Component);

    function TiBatteryCharge() {
        _classCallCheck(this, TiBatteryCharge);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiBatteryCharge).apply(this, arguments));
    }

    _createClass(TiBatteryCharge, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm8.333333333333334 16.666666666666668v10h18.333333333333336v-10h-18.333333333333336z m9.716666666666667 8.180000000000003l-2.0166666666666693-3.1800000000000033-4.366666666666664 0.7133333333333347 5.371666666666666-3.873333333333335 1.9583333333333321 3.16 4.33666666666667-0.716666666666665-5.283333333333331 3.8966666666666683z m13.616666666666667-8.180000000000003c0-2.7566666666666677-2.2433333333333323-5-5-5h-18.333333333333336c-2.756666666666665 0-4.999999999999998 2.243333333333334-4.999999999999998 5v10c0 2.7566666666666677 2.243333333333334 5 5 5h18.333333333333336c2.7566666666666677 0 5-2.2433333333333323 5-5 1.8400000000000034 0 3.3333333333333357-1.4933333333333323 3.3333333333333357-3.333333333333332v-3.333333333333332c0-1.8399999999999999-1.4933333333333323-3.333333333333332-3.333333333333332-3.333333333333332z m-3.333333333333332 10c0 0.9200000000000017-0.75 1.6666666666666679-1.6666666666666679 1.6666666666666679h-18.333333333333336c-0.9166666666666652 0-1.6666666666666652-0.7466666666666661-1.6666666666666652-1.6666666666666679v-10c0-0.9199999999999999 0.75-1.666666666666666 1.666666666666667-1.666666666666666h18.333333333333336c0.9166666666666679 0 1.6666666666666679 0.7466666666666661 1.6666666666666679 1.666666666666666v10z' })
                )
            );
        }
    }]);

    return TiBatteryCharge;
}(React.Component);

exports.default = TiBatteryCharge;
module.exports = exports['default'];