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

var MdUsb = function (_React$Component) {
    _inherits(MdUsb, _React$Component);

    function MdUsb() {
        _classCallCheck(this, MdUsb);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdUsb).apply(this, arguments));
    }

    _createClass(MdUsb, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm25 11.64h6.640000000000001v6.716666666666669h-1.6400000000000006v3.2833333333333314q0 1.4066666666666663-0.9766666666666666 2.383333333333333t-2.383333333333333 0.9766666666666666h-5v5.078333333333333q2.0333333333333314 1.0933333333333337 2.0333333333333314 3.2833333333333314 0 1.4833333333333343-1.0566666666666649 2.576666666666668t-2.616666666666667 1.0949999999999989-2.616666666666667-1.0949999999999989-1.0549999999999997-2.578333333333333q0-2.1883333333333326 2.0333333333333314-3.2833333333333314v-5.076666666666668h-5q-1.4083333333333332 0-2.383333333333333-0.9766666666666666t-0.9783333333333317-2.383333333333333v-3.4400000000000013q-2.0333333333333323-1.0899999999999999-2.0333333333333323-3.1999999999999993 0-1.5666666666666664 1.0949999999999998-2.616666666666667t2.578333333333333-1.0566666666666666 2.578333333333333 1.0549999999999997 1.0950000000000006 2.618333333333334q0 2.1866666666666674-1.9533333333333331 3.1999999999999993v3.4400000000000013h4.999999999999998v-13.283333333333333h-3.3599999999999994l5-6.716666666666668 5 6.716666666666668h-3.361666666666668v13.283333333333333h5v-3.2833333333333314h-1.6383333333333319v-6.716666666666669z' })
                )
            );
        }
    }]);

    return MdUsb;
}(React.Component);

exports.default = MdUsb;
module.exports = exports['default'];