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

var MdAirlineSeatLegroomReduced = function (_React$Component) {
    _inherits(MdAirlineSeatLegroomReduced, _React$Component);

    function MdAirlineSeatLegroomReduced() {
        _classCallCheck(this, MdAirlineSeatLegroomReduced);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdAirlineSeatLegroomReduced).apply(this, arguments));
    }

    _createClass(MdAirlineSeatLegroomReduced, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm8.360000000000001 20q0 2.0333333333333314 1.4833333333333325 3.5166666666666657t3.5166666666666657 1.4833333333333343h6.640000000000001v3.3599999999999994h-6.639999999999999q-3.4383333333333344 0-5.9-2.461666666666666t-2.460000000000001-5.898333333333333v-15h3.3599999999999994v15z m24.921666666666667 12.033333333333331q0.23333333333333428 1.1700000000000017-0.509999999999998 2.06666666666667t-1.913333333333334 0.8999999999999986h-7.5v-5l1.6416666666666657-6.640000000000001h-10q-2.033333333333333 0-3.5166666666666657-1.4833333333333343t-1.4833333333333343-3.518333333333331v-13.358333333333334h10v10h8.36q1.3283333333333331 0 2.3049999999999997 1.0166666666666657t0.9750000000000014 2.3416666666666686l-3.2816666666666663 11.641666666666666h2.344999999999999q0.9383333333333326 0 1.6799999999999997 0.5833333333333321t0.8999999999999986 1.446666666666669z' })
                )
            );
        }
    }]);

    return MdAirlineSeatLegroomReduced;
}(React.Component);

exports.default = MdAirlineSeatLegroomReduced;
module.exports = exports['default'];