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

var TiPi = function (_React$Component) {
    _inherits(TiPi, _React$Component);

    function TiPi() {
        _classCallCheck(this, TiPi);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiPi).apply(this, arguments));
    }

    _createClass(TiPi, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm31.178333333333335 14.225000000000001c-0.6499999999999986-0.6500000000000004-1.7049999999999983-0.6500000000000004-2.3566666666666656 0-2.1066666666666656 2.1066666666666656-5.533333333333335 2.1066666666666656-7.643333333333334 0-3.4083333333333314-3.4066666666666663-8.950000000000001-3.4033333333333324-12.356666666666667 0-0.6500000000000004 0.6500000000000004-0.6500000000000004 1.705 0 2.3566666666666656s1.705 0.6499999999999986 2.3566666666666674 0c0.6233333333333331-0.625 1.3666666666666671-1.041666666666666 2.1549999999999994-1.295v13.046666666666669c0 0.9216666666666669 0.7449999999999992 1.6666666666666679 1.666666666666666 1.6666666666666679s1.6666666666666679-0.745000000000001 1.6666666666666679-1.6666666666666679v-13.043333333333337c0.7866666666666653 0.25333333333333385 1.533333333333335 0.6666666666666661 2.155000000000001 1.2916666666666679 1.2800000000000011 1.2766666666666673 2.8583333333333343 2.073333333333334 4.511666666666667 2.3933333333333344v9.358333333333334c0 0.9216666666666669 0.745000000000001 1.6666666666666679 1.6666666666666679 1.6666666666666679s1.6666666666666679-0.745000000000001 1.6666666666666679-1.6666666666666679v-9.35666666666667c1.6533333333333324-0.31666666666666643 3.2333333333333343-1.1166666666666671 4.511666666666667-2.3949999999999996 0.6499999999999986-0.6500000000000004 0.6499999999999986-1.706666666666667 0-2.3583333333333343z' })
                )
            );
        }
    }]);

    return TiPi;
}(React.Component);

exports.default = TiPi;
module.exports = exports['default'];