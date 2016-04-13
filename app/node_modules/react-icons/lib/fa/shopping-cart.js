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

var FaShoppingCart = function (_React$Component) {
    _inherits(FaShoppingCart, _React$Component);

    function FaShoppingCart() {
        _classCallCheck(this, FaShoppingCart);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaShoppingCart).apply(this, arguments));
    }

    _createClass(FaShoppingCart, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm15.714285714285715 34.285714285714285q0 1.1600000000000037-0.8485714285714288 2.008571428571429t-2.008571428571429 0.8485714285714323-2.008571428571429-0.8485714285714252-0.8485714285714288-2.008571428571436 0.8485714285714288-2.008571428571429 2.008571428571429-0.8485714285714252 2.008571428571429 0.8485714285714252 0.8485714285714288 2.008571428571429z m20 0q0 1.1600000000000037-0.8485714285714252 2.008571428571429t-2.008571428571429 0.8485714285714323-2.008571428571429-0.8485714285714252-0.8485714285714323-2.008571428571436 0.8485714285714288-2.008571428571429 2.0085714285714324-0.8485714285714252 2.008571428571429 0.8485714285714252 0.8485714285714252 2.008571428571429z m2.857142857142854-24.285714285714285v11.42857142857143q0 0.5357142857142847-0.36857142857142833 0.9485714285714302t-0.9028571428571439 0.4799999999999969l-23.305714285714284 2.7228571428571406q0.2914285714285736 1.3400000000000034 0.2914285714285736 1.562857142857144 0 0.35714285714285765-0.5357142857142865 1.428571428571427h20.535714285714285q0.5799999999999983 0 1.0042857142857144 0.4242857142857126t0.42428571428571615 1.004285714285718-0.42428571428571615 1.0042857142857144-1.0042857142857144 0.42428571428571615h-22.857142857142854q-0.5800000000000018 0-1.0042857142857162-0.4242857142857126t-0.4242857142857144-1.004285714285718q0-0.24571428571428555 0.17857142857142883-0.7028571428571411t0.35714285714285765-0.8028571428571425 0.4800000000000004-0.8928571428571423 0.34714285714285786-0.6571428571428584l-3.9528571428571446-18.372857142857143h-4.5528571428571425q-0.5800000000000005-1.7763568394002505e-15-1.0042857142857147-0.42285714285714526t-0.4242857142857144-1.0057142857142845 0.42428571428571416-1.0000000000000009 1.0042857142857144-0.42857142857142794h5.7142857142857135q0.35714285714285765 0 0.6357142857142861 0.1471428571428568t0.43571428571428505 0.3457142857142861 0.28999999999999915 0.5471428571428572 0.17857142857142883 0.5800000000000001 0.12285714285714278 0.6571428571428575 0.09999999999999964 0.5814285714285718h26.808571428571433q0.5799999999999983 0 1.0042857142857144 0.42571428571428527t0.42428571428570905 1.001428571428571z' })
                )
            );
        }
    }]);

    return FaShoppingCart;
}(React.Component);

exports.default = FaShoppingCart;
module.exports = exports['default'];