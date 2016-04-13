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

var FaTelevision = function (_React$Component) {
    _inherits(FaTelevision, _React$Component);

    function FaTelevision() {
        _classCallCheck(this, FaTelevision);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaTelevision).apply(this, arguments));
    }

    _createClass(FaTelevision, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm37.33333333333333 27.333333333333332v-20q0-0.27066666666666706-0.19733333333333292-0.46933333333333316t-0.4693333333333314-0.19733333333333292h-33.33333333333333q-0.2706666666666666 0-0.46933333333333316 0.19733333333333292t-0.19733333333333603 0.46933333333333405v20q0 0.27066666666666706 0.19733333333333336 0.46933333333333493t0.46933333333333316 0.19733333333333292h33.33333333333333q0.2706666666666635 0 0.4693333333333314-0.19733333333333292t0.19733333333333292-0.46933333333333493z m2.6666666666666643-20v20q0 1.3733333333333348-0.978666666666669 2.3546666666666667t-2.3546666666666596 0.9786666666666655h-15.333333333333332v2.6666666666666643h7.333333333333332q0.293333333333333 0 0.4800000000000004 0.18666666666666742t0.18666666666666742 0.4799999999999969v1.3333333333333357q0 0.29333333333333655-0.18666666666666742 0.4799999999999969t-0.4800000000000004 0.18666666666666742h-17.333333333333332q-0.293333333333333 0-0.4800000000000004-0.18666666666666742t-0.18666666666666565-0.4799999999999969v-1.3333333333333357q0-0.29333333333333655 0.18666666666666742-0.4799999999999969t0.47999999999999865-0.18666666666666742h7.333333333333332v-2.666666666666668h-15.333333333333332q-1.3733333333333324 0-2.354666666666666-0.9786666666666655t-0.9786666666666664-2.354666666666663v-20q0-1.373333333333333 0.9786666666666666-2.3546666666666667t2.3546666666666667-0.9786666666666655h33.33333333333333q1.3733333333333348 0 2.3546666666666667 0.9786666666666664t0.978666666666669 2.3546666666666667z' })
                )
            );
        }
    }]);

    return FaTelevision;
}(React.Component);

exports.default = FaTelevision;
module.exports = exports['default'];