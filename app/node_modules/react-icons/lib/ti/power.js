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

var TiPower = function (_React$Component) {
    _inherits(TiPower, _React$Component);

    function TiPower() {
        _classCallCheck(this, TiPower);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiPower).apply(this, arguments));
    }

    _createClass(TiPower, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm19.166666666666668 30.955000000000002c-2.8933333333333344 0-5.613333333333333-1.1266666666666652-7.66-3.171666666666667-2.0450000000000017-2.046666666666667-3.173333333333334-4.766666666666666-3.173333333333334-7.661666666666665s1.1283333333333339-5.616666666666667 3.173333333333334-7.661666666666669c0.6500000000000004-0.6500000000000004 1.705-0.6500000000000004 2.3566666666666656 0s0.6500000000000004 1.705 0 2.3566666666666656c-1.416666666666666 1.4166666666666679-2.1966666666666654 3.3000000000000007-2.1966666666666654 5.305000000000003s0.7799999999999994 3.8883333333333354 2.196666666666667 5.305c1.416666666666666 1.4166666666666679 3.299999999999999 2.1950000000000003 5.303333333333333 2.1950000000000003s3.8866666666666667-0.7800000000000011 5.303333333333335-2.1950000000000003c1.4166666666666679-1.4166666666666679 2.1966666666666654-3.3000000000000007 2.1966666666666654-5.305s-0.7800000000000011-3.8883333333333354-2.1966666666666654-5.305c-0.6499999999999986-0.6500000000000004-0.6499999999999986-1.705 0-2.3566666666666656s1.7049999999999983-0.6500000000000004 2.3566666666666656 0c2.0450000000000017 2.0499999999999954 3.173333333333332 4.766666666666662 3.173333333333332 7.661666666666665s-1.1283333333333339 5.616666666666667-3.173333333333332 7.661666666666669c-2.046666666666667 2.0450000000000017-4.766666666666666 3.171666666666667-7.66 3.171666666666667z m0-12.621666666666666c-0.9216666666666669 0-1.6666666666666679-0.7466666666666661-1.6666666666666679-1.6666666666666679v-8.333333333333334c0-0.9199999999999999 0.745000000000001-1.666666666666667 1.6666666666666679-1.666666666666667s1.6666666666666679 0.746666666666667 1.6666666666666679 1.666666666666667v8.333333333333334c0 0.9200000000000017-0.745000000000001 1.6666666666666679-1.6666666666666679 1.6666666666666679z' })
                )
            );
        }
    }]);

    return TiPower;
}(React.Component);

exports.default = TiPower;
module.exports = exports['default'];