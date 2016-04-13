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

var FaLevelDown = function (_React$Component) {
    _inherits(FaLevelDown, _React$Component);

    function FaLevelDown() {
        _classCallCheck(this, FaLevelDown);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaLevelDown).apply(this, arguments));
    }

    _createClass(FaLevelDown, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm9.285714285714286 5.714285714285714h15.714285714285714q0.28999999999999915 0 0.5028571428571418 0.21142857142857174t0.21142857142856997 0.524285714285714v19.264285714285716h4.285714285714285q0.8928571428571423 0 1.2942857142857136 0.8242857142857147t-0.1999999999999993 1.5399999999999991l-7.142857142857142 8.57142857142857q-0.4028571428571439 0.4914285714285711-1.095714285714287 0.4914285714285711t-1.095714285714287-0.4928571428571402l-7.142857142857142-8.57142857142857q-0.5800000000000001-0.6914285714285704-0.1999999999999993-1.5399999999999991 0.40000000000000036-0.8257142857142838 1.2928571428571427-0.8257142857142838h4.2857142857142865v-14.282857142857148h-7.142857142857142q-0.31428571428571495 0-0.5571428571428569-0.24714285714285644l-3.571428571428571-4.285714285714286q-0.2914285714285718-0.31428571428571406-0.09142857142857075-0.7571428571428571 0.1999999999999993-0.42571428571428527 0.6471428571428568-0.42571428571428527z' })
                )
            );
        }
    }]);

    return FaLevelDown;
}(React.Component);

exports.default = FaLevelDown;
module.exports = exports['default'];