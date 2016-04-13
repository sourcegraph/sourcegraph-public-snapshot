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

var MdSatellite = function (_React$Component) {
    _inherits(MdSatellite, _React$Component);

    function MdSatellite() {
        _classCallCheck(this, MdSatellite);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdSatellite).apply(this, arguments));
    }

    _createClass(MdSatellite, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm8.360000000000001 30h23.28333333333334l-7.5-10-5.783333333333335 7.5-4.216666666666667-5z m0-10q4.843333333333334 0 8.241666666666665-3.4383333333333326t3.3983333333333334-8.283333333333335h-3.3599999999999994q0 3.4400000000000013-2.421666666666667 5.9t-5.8583333333333325 2.461666666666668v3.3599999999999994z m0-11.716666666666667v5.076666666666666q2.033333333333333 0 3.5166666666666675-1.5233333333333334t1.4833333333333343-3.5549999999999997h-5z m23.28-3.283333333333333q1.3283333333333402 0 2.3433333333333337 1.0166666666666666t1.0166666666666657 2.3400000000000007v23.28333333333333q0 1.326666666666668-1.0166666666666657 2.3416666666666686t-2.3433333333333337 1.0166666666666657h-23.28333333333333q-1.3266666666666689 0-2.3416666666666686-1.0166666666666657t-1.0150000000000006-2.341666666666665v-23.28333333333334q0-1.3266666666666653 1.0166666666666666-2.341666666666665t2.3400000000000007-1.0150000000000006h23.28333333333333z' })
                )
            );
        }
    }]);

    return MdSatellite;
}(React.Component);

exports.default = MdSatellite;
module.exports = exports['default'];