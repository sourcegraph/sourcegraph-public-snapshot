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

var FaExchange = function (_React$Component) {
    _inherits(FaExchange, _React$Component);

    function FaExchange() {
        _classCallCheck(this, FaExchange);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaExchange).apply(this, arguments));
    }

    _createClass(FaExchange, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm40 26.42857142857143v4.285714285714285q0 0.28999999999999915-0.21142857142856997 0.5028571428571418t-0.5028571428571453 0.21142857142857352h-30.714285714285715v4.285714285714285q1.7763568394002505e-15 0.28999999999999915-0.21142857142856997 0.5028571428571453t-0.5028571428571418 0.21142857142856997q-0.2671428571428569 0-0.5357142857142856-0.2228571428571442l-7.12-7.142857142857142q-0.20142857142857196-0.1999999999999993-0.20142857142857196-0.4914285714285711 0-0.31428571428571317 0.2-0.514285714285716l7.142857142857143-7.142857142857142q0.20285714285714285-0.1999999999999993 0.5142857142857142-0.1999999999999993 0.28999999999999915 0 0.5028571428571436 0.21142857142856997t0.21142857142856997 0.5028571428571453v4.285714285714285h30.714285714285715q0.28999999999999915 0 0.5028571428571453 0.21142857142856997t0.21142857142856997 0.5028571428571453z m0-12.142857142857142q0 0.31428571428571495-0.20000000000000284 0.5142857142857142l-7.142857142857146 7.142857142857144q-0.20285714285714107 0.1999999999999993-0.5142857142857125 0.1999999999999993-0.28999999999999915 0-0.5028571428571418-0.21142857142856997t-0.21142857142856641-0.5028571428571453v-4.285714285714285h-30.714285714285715q-0.29000000000000103 0-0.5028571428571439-0.21142857142856997t-0.2114285714285714-0.5028571428571453v-4.285714285714285q0-0.28999999999999915 0.2114285714285714-0.5028571428571436t0.5028571428571429-0.21142857142857352h30.714285714285715v-4.285714285714286q0-0.31428571428571406 0.1999999999999993-0.5142857142857142t0.514285714285716-0.20000000000000018q0.2671428571428578 0 0.5357142857142847 0.22285714285714242l7.121428571428567 7.120000000000001q0.20000000000000284 0.1999999999999993 0.20000000000000284 0.5142857142857142z' })
                )
            );
        }
    }]);

    return FaExchange;
}(React.Component);

exports.default = FaExchange;
module.exports = exports['default'];