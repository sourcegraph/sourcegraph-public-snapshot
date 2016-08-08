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

var MdPhonelinkRing = function (_React$Component) {
    _inherits(MdPhonelinkRing, _React$Component);

    function MdPhonelinkRing() {
        _classCallCheck(this, MdPhonelinkRing);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdPhonelinkRing).apply(this, arguments));
    }

    _createClass(MdPhonelinkRing, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm23.36 33.36v-26.716666666666665h-16.716666666666665v26.716666666666665h16.716666666666665z m0-31.720000000000002q1.3283333333333331 0 2.3049999999999997 1.0166666666666666t0.9750000000000014 2.341666666666667v29.999999999999996q0 1.326666666666668-0.9766666666666666 2.3416666666666686t-2.306666666666665 1.0166666666666657h-16.713333333333335q-1.330000000000001 0-2.3066666666666675-1.0166666666666657t-0.9766666666666666-2.3399999999999963v-30q0-1.33 0.9766666666666666-2.345t2.3050000000000006-1.0166666666666666h16.71666666666667z m6.640000000000001 14.68833333333334q1.5633333333333326 1.6400000000000006 1.5633333333333326 3.671666666666667t-1.5633333333333326 3.516666666666662l-1.6400000000000006-1.7199999999999989q1.4066666666666663-1.9533333333333331 0-3.828333333333333z m3.5166666666666657-3.5166666666666657q3.123333333333335 2.9700000000000006 3.123333333333335 7.150000000000002t-3.123333333333335 7.071666666666658l-1.7199999999999989-1.7199999999999989q2.2666666666666657-2.421666666666667 2.2666666666666657-5.508333333333333t-2.2666666666666657-5.273333333333333z' })
                )
            );
        }
    }]);

    return MdPhonelinkRing;
}(React.Component);

exports.default = MdPhonelinkRing;
module.exports = exports['default'];