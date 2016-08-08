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

var FaMinus = function (_React$Component) {
    _inherits(FaMinus, _React$Component);

    function FaMinus() {
        _classCallCheck(this, FaMinus);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaMinus).apply(this, arguments));
    }

    _createClass(FaMinus, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm35.714285714285715 16.42857142857143v4.285714285714285q0 0.8928571428571423-0.6257142857142881 1.5171428571428578t-1.5171428571428578 0.6257142857142846h-27.142857142857142q-0.8928571428571432 0-1.5171428571428578-0.6257142857142846t-0.6257142857142837-1.5171428571428578v-4.285714285714285q0-0.8928571428571423 0.6257142857142854-1.5171428571428578t1.5171428571428578-0.6257142857142863h27.14285714285714q0.8928571428571459 0 1.5171428571428578 0.6257142857142863t0.6257142857142881 1.5171428571428578z' })
                )
            );
        }
    }]);

    return FaMinus;
}(React.Component);

exports.default = FaMinus;
module.exports = exports['default'];