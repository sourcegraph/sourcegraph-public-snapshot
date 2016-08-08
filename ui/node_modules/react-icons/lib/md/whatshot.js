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

var MdWhatshot = function (_React$Component) {
    _inherits(MdWhatshot, _React$Component);

    function MdWhatshot() {
        _classCallCheck(this, MdWhatshot);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdWhatshot).apply(this, arguments));
    }

    _createClass(MdWhatshot, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm19.533333333333335 31.640000000000004q3.3583333333333343 0 5.661666666666665-2.3049999999999997t2.3049999999999997-5.661666666666665q0-3.4383333333333326-0.9383333333333326-6.716666666666669-2.421666666666667 3.1999999999999993-7.733333333333334 4.295000000000002-4.690000000000001 1.0166666666666657-4.690000000000001 5.156666666666666 0 2.1883333333333326 1.5633333333333326 3.711666666666666t3.830000000000002 1.5249999999999986z m2.966666666666665-30.54666666666667q5 4.0633333333333335 7.93 9.883333333333333t2.9299999999999997 12.383333333333333q0 5.466666666666669-3.9066666666666663 9.375t-9.453333333333333 3.9083333333333385-9.453333333333333-3.9100000000000037-3.9066666666666663-9.373333333333335q0-8.438333333333333 5.390000000000001-14.843333333333335v0.625q0 2.578333333333333 1.7166666666666668 4.375t4.299999999999999 1.7966666666666669q2.5 0 4.100000000000001-1.7583333333333329t1.6000000000000014-4.413333333333334q0-1.5633333333333335-0.3116666666666674-3.5933333333333337t-0.6233333333333348-3.205z' })
                )
            );
        }
    }]);

    return MdWhatshot;
}(React.Component);

exports.default = MdWhatshot;
module.exports = exports['default'];