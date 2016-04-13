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

var TiArrowMove = function (_React$Component) {
    _inherits(TiArrowMove, _React$Component);

    function TiArrowMove() {
        _classCallCheck(this, TiArrowMove);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiArrowMove).apply(this, arguments));
    }

    _createClass(TiArrowMove, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm29.51166666666667 13.821666666666665c-0.6499999999999986-0.6500000000000004-1.7049999999999983-0.6500000000000004-2.3566666666666656 0s-0.6499999999999986 1.705 0 2.3566666666666656l2.154999999999994 2.1550000000000047h-7.643333333333331v-7.643333333333336l2.155000000000001 2.1549999999999994c0.3249999999999993 0.3249999999999993 0.75 0.4883333333333333 1.1783333333333346 0.4883333333333333s0.8533333333333317-0.163333333333334 1.1783333333333346-0.4883333333333333c0.6499999999999986-0.6500000000000004 0.6499999999999986-1.705 0-2.3566666666666656l-6.178333333333338-6.178333333333334-6.178333333333335 6.178333333333334c-0.6500000000000004 0.6500000000000004-0.6500000000000004 1.705 0 2.3566666666666656s1.705 0.6500000000000004 2.3566666666666656 0l2.1550000000000047-2.1549999999999994v7.643333333333336h-7.643333333333336l2.1549999999999994-2.155000000000001c0.6500000000000004-0.6500000000000004 0.6500000000000004-1.705 0-2.3566666666666656s-1.705-0.6500000000000004-2.3566666666666656 0l-6.178333333333334 6.178333333333331 6.178333333333334 6.178333333333335c0.3249999999999993 0.3249999999999993 0.75 0.4883333333333333 1.1783333333333328 0.4883333333333333s0.8533333333333335-0.163333333333334 1.1783333333333328-0.4883333333333333c0.6500000000000004-0.6499999999999986 0.6500000000000004-1.7049999999999983 0-2.3566666666666656l-2.1549999999999994-2.155000000000001h7.643333333333336v7.643333333333334l-2.155000000000001-2.155000000000001c-0.6500000000000004-0.6499999999999986-1.705-0.6499999999999986-2.3566666666666656 0s-0.6500000000000004 1.7049999999999983 0 2.3566666666666656l6.178333333333331 6.178333333333338 6.178333333333335-6.178333333333335c0.6499999999999986-0.6499999999999986 0.6499999999999986-1.7049999999999983 0-2.3566666666666656s-1.7049999999999983-0.6499999999999986-2.3566666666666656 0l-2.155000000000001 2.154999999999994v-7.643333333333331h7.643333333333334l-2.155000000000001 2.155000000000001c-0.6499999999999986 0.6499999999999986-0.6499999999999986 1.7049999999999983 0 2.3566666666666656 0.3249999999999993 0.3249999999999993 0.75 0.4883333333333333 1.1783333333333346 0.4883333333333333s0.8533333333333317-0.163333333333334 1.1783333333333346-0.4883333333333333l6.178333333333335-6.178333333333335-6.178333333333335-6.178333333333335z' })
                )
            );
        }
    }]);

    return TiArrowMove;
}(React.Component);

exports.default = TiArrowMove;
module.exports = exports['default'];