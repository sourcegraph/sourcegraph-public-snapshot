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

var FaArrowsAlt = function (_React$Component) {
    _inherits(FaArrowsAlt, _React$Component);

    function FaArrowsAlt() {
        _classCallCheck(this, FaArrowsAlt);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaArrowsAlt).apply(this, arguments));
    }

    _createClass(FaArrowsAlt, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm31.495714285714286 12.075714285714286l-7.924285714285713 7.924285714285714 7.924285714285713 7.924285714285716 3.2142857142857153-3.2142857142857153q0.6471428571428604-0.6914285714285704 1.5628571428571405-0.31428571428571317 0.8714285714285737 0.38142857142857167 0.8714285714285737 1.3185714285714276v10q0 0.5799999999999983-0.42428571428571615 1.0042857142857144t-1.0057142857142836 0.42428571428571615h-10q-0.937142857142856 0-1.3171428571428585-0.8928571428571459-0.379999999999999-0.8714285714285737 0.31428571428571317-1.5399999999999991l3.2142857142857153-3.2142857142857153-7.925714285714285-7.924285714285713-7.924285714285714 7.924285714285713 3.2142857142857135 3.2142857142857153q0.6914285714285722 0.6714285714285708 0.31428571428571495 1.5399999999999991-0.38285714285714256 0.8928571428571459-1.3185714285714276 0.8928571428571459h-10q-0.580000000000001 0-1.0042857142857153-0.42428571428571615t-0.42428571428571393-1.0042857142857144v-10q0-0.937142857142856 0.8928571428571428-1.3171428571428585 0.8714285714285719-0.379999999999999 1.54 0.31428571428571317l3.2142857142857144 3.2142857142857153 7.924285714285716-7.925714285714285-7.924285714285716-7.924285714285714-3.2142857142857144 3.2142857142857135q-0.4242857142857144 0.42428571428571615-1.0042857142857144 0.42428571428571615-0.2671428571428569 0-0.5357142857142856-0.1114285714285721-0.8928571428571428-0.379999999999999-0.8928571428571428-1.3171428571428567v-10q0-0.580000000000001 0.4242857142857144-1.0042857142857153t1.004285714285714-0.42428571428571393h10q0.9371428571428577 0 1.3171428571428567 0.8928571428571428 0.3800000000000008 0.8714285714285719-0.31428571428571495 1.54l-3.2142857142857153 3.2142857142857144 7.925714285714289 7.924285714285716 7.924285714285716-7.924285714285714-3.2142857142857153-3.2142857142857144q-0.6914285714285704-0.6714285714285717-0.31428571428571317-1.54 0.3828571428571408-0.8928571428571446 1.3185714285714276-0.8928571428571446h10q0.5799999999999983 0 1.0042857142857144 0.4242857142857144t0.42428571428571615 1.004285714285714v10q0 0.9371428571428577-0.8714285714285737 1.3171428571428567-0.2857142857142847 0.11142857142857387-0.5571428571428569 0.11142857142857387-0.5799999999999983 0-1.0042857142857144-0.4242857142857144z' })
                )
            );
        }
    }]);

    return FaArrowsAlt;
}(React.Component);

exports.default = FaArrowsAlt;
module.exports = exports['default'];