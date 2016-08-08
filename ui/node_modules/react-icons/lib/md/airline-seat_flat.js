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

var MdAirlineSeatFlat = function (_React$Component) {
    _inherits(MdAirlineSeatFlat, _React$Component);

    function MdAirlineSeatFlat() {
        _classCallCheck(this, MdAirlineSeatFlat);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdAirlineSeatFlat).apply(this, arguments));
    }

    _createClass(MdAirlineSeatFlat, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm11.875 20.156666666666666q-1.4833333333333343 1.4833333333333343-3.5166666666666657 1.5233333333333334t-3.5133333333333336-1.4450000000000003-1.5233333333333334-3.5133333333333354 1.4450000000000003-3.5166666666666657 3.5166666666666675-1.5216666666666665 3.5133333333333336 1.4433333333333334 1.5233333333333334 3.5166666666666657-1.4450000000000003 3.5133333333333354z m-8.516666666666667 3.203333333333333h33.28333333333333v3.2833333333333314h-10v3.356666666666669h-13.283333333333333v-3.3583333333333343h-10v-3.2833333333333314z m33.28333333333333-5v3.2833333333333314h-21.641666666666666v-10h15q2.7333333333333343 0 4.688333333333333 1.9900000000000002t1.9533333333333331 4.726666666666668z' })
                )
            );
        }
    }]);

    return MdAirlineSeatFlat;
}(React.Component);

exports.default = MdAirlineSeatFlat;
module.exports = exports['default'];