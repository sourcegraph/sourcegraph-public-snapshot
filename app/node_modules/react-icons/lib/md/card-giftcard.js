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

var MdCardGiftcard = function (_React$Component) {
    _inherits(MdCardGiftcard, _React$Component);

    function MdCardGiftcard() {
        _classCallCheck(this, MdCardGiftcard);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdCardGiftcard).apply(this, arguments));
    }

    _createClass(MdCardGiftcard, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm33.36 23.36v-10h-8.516666666666666l3.5166666666666657 4.688333333333333-2.7333333333333343 1.9499999999999993q-5-6.795-5.626666666666665-7.655000000000001-0.625 0.8599999999999994-5.625 7.656666666666666l-2.7333333333333343-1.9499999999999993 3.5133333333333336-4.690000000000001h-8.516666666666667v10h26.720000000000002z m0 8.280000000000001v-3.2833333333333314h-26.716666666666665v3.2833333333333314h26.716666666666665z m-18.36-25q-0.7033333333333331 0-1.1716666666666669 0.5083333333333329t-0.4666666666666668 1.2100000000000009 0.4666666666666668 1.1716666666666669 1.1716666666666669 0.46999999999999886 1.1716666666666669-0.4666666666666668 0.466666666666665-1.1733333333333338-0.466666666666665-1.211666666666667-1.1716666666666669-0.5099999999999989z m10 0q-0.7033333333333331 0-1.1716666666666669 0.5083333333333329t-0.466666666666665 1.2100000000000009 0.466666666666665 1.1716666666666669 1.1716666666666669 0.46999999999999886 1.1716666666666669-0.4666666666666668 0.466666666666665-1.1733333333333338-0.466666666666665-1.211666666666667-1.1716666666666669-0.5099999999999989z m8.36 3.3599999999999994q1.4066666666666663 0 2.3433333333333337 0.9766666666666666t0.9383333333333326 2.383333333333333v18.283333333333335q0 1.4049999999999976-0.9383333333333326 2.3833333333333364t-2.3433333333333337 0.9733333333333292h-26.716666666666665q-1.408333333333334 0-2.3450000000000006-0.9750000000000014t-0.94-2.383333333333333v-18.28333333333333q0-1.4049999999999994 0.938333333333333-2.383333333333333t2.3433333333333337-0.9750000000000014h3.671666666666667q-0.3116666666666674-1.093333333333332-0.3116666666666674-1.6399999999999988 0-2.033333333333333 1.4833333333333343-3.5166666666666666t3.5166666666666657-1.4816666666666678q2.578333333333333 0 4.140000000000001 2.188333333333334l0.8599999999999994 1.166666666666667 0.8599999999999994-1.1716666666666669q1.5633333333333361-2.186666666666667 4.140000000000001-2.186666666666667 2.0333333333333314 0 3.5166666666666657 1.483333333333333t1.4833333333333343 3.5166666666666675q0 0.5466666666666669-0.3133333333333326 1.6400000000000006h3.671666666666667z' })
                )
            );
        }
    }]);

    return MdCardGiftcard;
}(React.Component);

exports.default = MdCardGiftcard;
module.exports = exports['default'];