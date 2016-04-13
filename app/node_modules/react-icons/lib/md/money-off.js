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

var MdMoneyOff = function (_React$Component) {
    _inherits(MdMoneyOff, _React$Component);

    function MdMoneyOff() {
        _classCallCheck(this, MdMoneyOff);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdMoneyOff).apply(this, arguments));
    }

    _createClass(MdMoneyOff, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm8.906666666666668 6.796666666666668l24.296666666666667 24.375-2.1099999999999994 2.1099999999999994-3.671666666666667-3.75q-1.4833333333333343 1.326666666666668-4.063333333333333 1.8733333333333348v3.594999999999999h-5v-3.5933333333333337q-2.578333333333333-0.5466666666666669-4.296666666666667-2.1883333333333326t-1.875-4.218333333333334h3.671666666666667q0.3133333333333326 3.5166666666666657 5 3.5166666666666657 2.8900000000000006 0 3.9833333333333343-1.5666666666666664l-5.858333333333334-5.780000000000001q-6.483333333333334-1.9533333333333331-6.483333333333334-6.563333333333333l-5.706666666666667-5.703333333333333z m11.953333333333331 4.6883333333333335q-1.4833333333333343 0-2.578333333333333 0.47000000000000064l-2.423333333333332-2.4216666666666686q1.0933333333333337-0.5500000000000007 2.5-0.9399999999999995v-3.5933333333333337h5v3.671666666666667q2.5 0.625 3.866666666666667 2.3433333333333337t1.4466666666666654 3.9849999999999994h-3.671666666666667q-0.1566666666666663-3.5166666666666657-4.140000000000001-3.5166666666666657z' })
                )
            );
        }
    }]);

    return MdMoneyOff;
}(React.Component);

exports.default = MdMoneyOff;
module.exports = exports['default'];