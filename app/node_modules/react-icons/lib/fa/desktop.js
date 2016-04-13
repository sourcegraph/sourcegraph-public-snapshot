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

var FaDesktop = function (_React$Component) {
    _inherits(FaDesktop, _React$Component);

    function FaDesktop() {
        _classCallCheck(this, FaDesktop);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaDesktop).apply(this, arguments));
    }

    _createClass(FaDesktop, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm37.33333333333333 22v-17.333333333333332q0-0.27066666666666706-0.19733333333333292-0.46933333333333316t-0.4693333333333314-0.1973333333333347h-33.33333333333333q-0.2706666666666666 0-0.46933333333333316 0.19733333333333292t-0.19733333333333603 0.46933333333333316v17.333333333333336q0 0.27066666666666706 0.19733333333333336 0.46933333333333493t0.46933333333333316 0.19733333333332936h33.33333333333333q0.2706666666666635 0 0.4693333333333314-0.19733333333333292t0.19733333333333292-0.4693333333333314z m2.6666666666666643-17.333333333333332v22.666666666666664q0 1.3733333333333348-0.978666666666669 2.3546666666666667t-2.3546666666666596 0.9786666666666655h-11.333333333333332q0 0.770666666666667 0.33333333333333215 1.6133333333333368t0.6666666666666679 1.4799999999999969 0.33333333333333215 0.9066666666666663q0 0.5413333333333341-0.3960000000000008 0.9373333333333349t-0.9373333333333314 0.3960000000000008h-10.666666666666666q-0.5413333333333341 0-0.9373333333333331-0.3960000000000008t-0.3960000000000008-0.9373333333333349q0-0.29333333333333655 0.3333333333333339-0.9173333333333318t0.6666666666666661-1.458666666666666 0.3333333333333339-1.6266666666666652h-11.333333333333332q-1.3733333333333342 0-2.3546666666666676-0.977333333333334t-0.9786666666666664-2.3533333333333353v-22.666666666666664q0-1.3773333333333344 0.9786666666666666-2.357333333333335t2.3546666666666667-0.9759999999999998h33.33333333333333q1.3733333333333348 0 2.3546666666666667 0.9773333333333334t0.978666666666669 2.3559999999999994z' })
                )
            );
        }
    }]);

    return FaDesktop;
}(React.Component);

exports.default = FaDesktop;
module.exports = exports['default'];