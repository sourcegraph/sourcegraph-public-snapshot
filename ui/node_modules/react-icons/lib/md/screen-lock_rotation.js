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

var MdScreenLockRotation = function (_React$Component) {
    _inherits(MdScreenLockRotation, _React$Component);

    function MdScreenLockRotation() {
        _classCallCheck(this, MdScreenLockRotation);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdScreenLockRotation).apply(this, arguments));
    }

    _createClass(MdScreenLockRotation, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm27.96666666666667 4.140000000000001v0.8599999999999994h5.704999999999998v-0.8600000000000003q0-1.1716666666666669-0.8200000000000003-1.9916666666666671t-1.9899999999999984-0.8216666666666657-2.033333333333335 0.8200000000000001-0.8583333333333343 1.9899999999999998z m-1.3250000000000028 10.86q-0.7033333333333331 0-1.1716666666666669-0.4666666666666668t-0.46999999999999886-1.174999999999999v-6.716666666666668q0-0.7050000000000001 0.466666666666665-1.173333333333333t1.173333333333332-0.4666666666666668v-0.8616666666666664q0-1.7166666666666668 1.25-2.93t2.9666666666666686-1.2100000000000004 2.9333333333333336 1.2116666666666667 1.2100000000000009 2.9299999999999997v0.8583333333333334q0.7033333333333331 0 1.1716666666666669 0.46999999999999975t0.46666666666666856 1.1716666666666669v6.716666666666668q0 0.7050000000000001-0.46666666666666856 1.1733333333333338t-1.1716666666666669 0.46833333333333194h-8.36z m-12.5 19.14l2.1883333333333326-2.1883333333333326 6.328333333333333 6.329999999999998-1.0933333333333337 0.07833333333333314q-7.813333333333333 0-13.555-5.313333333333333t-6.368333333333331-13.046666666666667h2.5q0.4666666666666668 4.608333333333334 3.163333333333333 8.396666666666668t6.836666666666668 5.741666666666667z m24.608333333333334-12.89q0.7833333333333314 0.7033333333333331 0.7833333333333314 1.7583333333333329t-0.7833333333333314 1.836666666666666l-10.625 10.546666666666667q-0.7033333333333331 0.7833333333333314-1.7166666666666686 0.7833333333333314t-1.8000000000000007-0.7833333333333314l-20-20q-0.7800000000000002-0.7033333333333331-0.7800000000000002-1.7166666666666668t0.7833333333333332-1.8000000000000007l10.543333333333337-10.621666666666664q0.7050000000000001-0.7816666666666672 1.7616666666666667-0.7816666666666672t1.8333333333333321 0.7833333333333332l4.066666666666666 4.0616666666666665-2.344999999999999 2.341666666666667-3.5166666666666657-3.4383333333333344-9.450000000000001 9.375 18.903333333333336 18.905 9.374999999999993-9.450000000000003-3.6700000000000017-3.673333333333332 2.3433333333333337-2.3433333333333337z' })
                )
            );
        }
    }]);

    return MdScreenLockRotation;
}(React.Component);

exports.default = MdScreenLockRotation;
module.exports = exports['default'];