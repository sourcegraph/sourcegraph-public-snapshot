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

var MdAirlineSeatLegroomNormal = function (_React$Component) {
    _inherits(MdAirlineSeatLegroomNormal, _React$Component);

    function MdAirlineSeatLegroomNormal() {
        _classCallCheck(this, MdAirlineSeatLegroomNormal);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdAirlineSeatLegroomNormal).apply(this, arguments));
    }

    _createClass(MdAirlineSeatLegroomNormal, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm34.14000000000001 30q1.0166666666666657 0 1.7583333333333329 0.7033333333333331t0.740000000000002 1.7966666666666669-0.7416666666666671 1.7966666666666669-1.759999999999998 0.7033333333333331h-7.5v-11.64h-11.636666666666677q-2.033333333333333 0-3.5166666666666657-1.4833333333333343t-1.4833333333333343-3.518333333333331v-13.358333333333334h10v10h8.36q1.3283333333333331 0 2.3049999999999997 1.0166666666666657t0.9750000000000014 2.3416666666666686v11.641666666666666h2.5z m-25.78000000000001-10q1.7763568394002505e-15 2.0333333333333314 1.4833333333333343 3.5166666666666657t3.5166666666666657 1.4833333333333343h10v3.3599999999999994h-10q-3.4383333333333344 0-5.9-2.461666666666666t-2.459999999999999-5.898333333333333v-15h3.3599999999999994v15z' })
                )
            );
        }
    }]);

    return MdAirlineSeatLegroomNormal;
}(React.Component);

exports.default = MdAirlineSeatLegroomNormal;
module.exports = exports['default'];