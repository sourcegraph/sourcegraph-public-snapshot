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

var MdUnarchive = function (_React$Component) {
    _inherits(MdUnarchive, _React$Component);

    function MdUnarchive() {
        _classCallCheck(this, MdUnarchive);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdUnarchive).apply(this, arguments));
    }

    _createClass(MdUnarchive, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm8.516666666666667 8.360000000000001h22.96666666666667l-1.5633333333333326-1.7166666666666668h-20z m11.483333333333333 7.5l-9.139999999999999 9.139999999999999h5.783333333333333v3.3599999999999994h6.716666666666669v-3.3599999999999994h5.783333333333335z m14.216666666666669-7.188333333333334q0.7833333333333314 0.9383333333333326 0.7833333333333314 2.1883333333333326v20.783333333333335q0 1.3266666666666644-1.0166666666666657 2.341666666666665t-2.3416666666666686 1.0166666666666657h-23.28333333333333q-1.405000000000002 0-2.3833333333333346-1.0166666666666657t-0.9749999999999996-2.341666666666665v-20.785000000000004q0-1.2499999999999982 0.7833333333333332-2.1866666666666656l2.2633333333333336-2.7333333333333334q0.7833333333333332-0.9400000000000004 1.9533333333333331-0.9400000000000004h20q1.1716666666666669 0 1.9533333333333331 0.9366666666666665z' })
                )
            );
        }
    }]);

    return MdUnarchive;
}(React.Component);

exports.default = MdUnarchive;
module.exports = exports['default'];