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

var MdMotorcycle = function (_React$Component) {
    _inherits(MdMotorcycle, _React$Component);

    function MdMotorcycle() {
        _classCallCheck(this, MdMotorcycle);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdMotorcycle).apply(this, arguments));
    }

    _createClass(MdMotorcycle, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm31.640000000000004 28.36q2.033333333333335 0 3.5166666666666693-1.4833333333333343t1.4833333333333343-3.5166666666666657-1.4833333333333343-3.5166666666666657-3.5166666666666657-1.4833333333333343-3.5166666666666657 1.4833333333333343-1.4833333333333343 3.5166666666666657 1.4833333333333343 3.5166666666666657 3.5166666666666657 1.4833333333333343z m-18.593333333333337-3.3599999999999994h-4.688333333333334v-3.3599999999999994h4.688333333333334q-0.5466666666666669-1.4833333333333343-1.836666666666666-2.383333333333333t-2.8499999999999996-0.8999999999999986q-2.033333333333333 0-3.5166666666666666 1.4866666666666681t-1.4833333333333334 3.5166666666666657 1.4833333333333334 3.513333333333332 3.5166666666666666 1.4833333333333343q1.5616666666666674 0 2.8499999999999996-0.9366666666666674t1.836666666666666-2.4200000000000017z m19.375-9.921666666666667q3.203333333333333 0.2333333333333325 5.390000000000001 2.578333333333333t2.1883333333333326 5.703333333333333q0 3.5166666666666657-2.421666666666667 5.899999999999999t-5.938333333333333 2.383333333333333-5.899999999999999-2.383333333333333-2.383333333333333-5.899999999999999q0-1.5633333333333326 0.46999999999999886-2.9666666666666686l-4.610000000000003 4.606666666666669h-2.7333333333333307q-0.6266666666666669 2.8916666666666657-2.8533333333333335 4.766666666666666t-5.273333333333333 1.875q-3.5166666666666666 0-5.9383333333333335-2.383333333333333t-2.418333333333334-5.899999999999999 2.421666666666667-5.936666666666667 5.936666666666666-2.421666666666667h19.296666666666667l-3.3599999999999994-3.3599999999999994h-5.938333333333333v-3.283333333333333h7.343333333333334z' })
                )
            );
        }
    }]);

    return MdMotorcycle;
}(React.Component);

exports.default = MdMotorcycle;
module.exports = exports['default'];