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

var MdExposurePlus2 = function (_React$Component) {
    _inherits(MdExposurePlus2, _React$Component);

    function MdExposurePlus2() {
        _classCallCheck(this, MdExposurePlus2);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdExposurePlus2).apply(this, arguments));
    }

    _createClass(MdExposurePlus2, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm13.360000000000001 11.64v6.716666666666669h6.639999999999999v3.2833333333333314h-6.639999999999999v6.716666666666669h-3.360000000000001v-6.716666666666669h-6.64v-3.2833333333333314h6.64v-6.716666666666669h3.3599999999999994z m13.358333333333333 15.55h9.923333333333332v2.8099999999999987h-14.375v-2.5l6.949999999999999-7.576666666666668q1.408333333333335-1.4066666666666663 2.423333333333332-3.125 0.6250000000000036-1.0166666666666657 0.6250000000000036-2.1900000000000013 0-1.25-0.8616666666666681-2.421666666666667-0.783333333333335-1.0166666666666657-2.3433333333333337-1.0166666666666657-1.6400000000000006 0-2.7333333333333343 1.0950000000000006-0.8616666666666681 0.8599999999999994-0.8616666666666681 2.7333333333333325h-3.5933333333333337q0-2.8116666666666674 1.875-4.686666666666667 1.0916666666666686-1.0933333333333337 2.2633333333333354-1.4833333333333343 1.4066666666666663-0.47000000000000064 3.125-0.47000000000000064 1.4066666666666663 0 2.8133333333333326 0.39000000000000057 1.56666666666667 0.625 2.1116666666666646 1.1716666666666669 1.7966666666666669 1.5633333333333326 1.7966666666666669 4.296666666666667 0 1.876666666666667-1.25 3.908333333333333-0.9383333333333326 1.5633333333333326-1.3283333333333331 1.9533333333333331-1.0933333333333337 1.25-1.7966666666666669 1.9533333333333331z' })
                )
            );
        }
    }]);

    return MdExposurePlus2;
}(React.Component);

exports.default = MdExposurePlus2;
module.exports = exports['default'];