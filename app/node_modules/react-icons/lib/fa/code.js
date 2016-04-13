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

var FaCode = function (_React$Component) {
    _inherits(FaCode, _React$Component);

    function FaCode() {
        _classCallCheck(this, FaCode);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaCode).apply(this, arguments));
    }

    _createClass(FaCode, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm12.853333333333333 30.479999999999997l-1.040000000000001 1.0399999999999991q-0.20800000000000018 0.2079999999999984-0.4800000000000004 0.2079999999999984t-0.47733333333333405-0.2079999999999984l-9.706666666666667-9.706666666666667q-0.21066666666666456-0.21066666666666478-0.21066666666666456-0.4799999999999969t0.20799999999999996-0.4800000000000004l9.709333333333333-9.706666666666667q0.20800000000000018-0.20800000000000018 0.4800000000000004-0.20800000000000018t0.47733333333333405 0.20800000000000018l1.040000000000001 1.0399999999999991q0.20933333333333337 0.20933333333333337 0.20933333333333337 0.4800000000000004t-0.20800000000000018 0.4800000000000004l-8.188000000000002 8.186666666666667 8.186666666666666 8.186666666666664q0.20933333333333337 0.20933333333333337 0.20933333333333337 0.4800000000000004t-0.20800000000000018 0.4800000000000004z m12.31333333333333-22.230666666666664l-7.7706666666666635 26.89733333333333q-0.08399999999999963 0.26933333333333564-0.32266666666666666 0.4053333333333313t-0.4893333333333345 0.053333333333334565l-1.293333333333333-0.35600000000000165q-0.26933333333333387-0.08400000000000318-0.4053333333333331-0.3226666666666631t-0.05333333333333279-0.5106666666666655l7.773333333333333-26.896q0.08266666666666822-0.27066666666666706 0.3200000000000003-0.4066666666666663t0.4906666666666659-0.05333333333333368l1.293333333333333 0.35599999999999987q0.2693333333333321 0.08399999999999963 0.4053333333333349 0.32266666666666666t0.053333333333334565 0.5106666666666664z m13.686666666666667 13.564l-9.706666666666667 9.706666666666667q-0.2079999999999984 0.2079999999999984-0.4800000000000004 0.2079999999999984t-0.47733333333333405-0.2079999999999984l-1.0399999999999991-1.0399999999999991q-0.20933333333333337-0.20933333333333337-0.20933333333333337-0.4800000000000004t0.2079999999999984-0.4800000000000004l8.189333333333337-8.186666666666667-8.186666666666667-8.186666666666666q-0.20933333333333337-0.20933333333333337-0.20933333333333337-0.4800000000000004t0.2079999999999984-0.4800000000000004l1.0426666666666655-1.040000000000001q0.2079999999999984-0.20800000000000018 0.4800000000000004-0.20800000000000018t0.47733333333333405 0.20800000000000018l9.706666666666663 9.706666666666667q0.20933333333333337 0.20933333333333337 0.20933333333333337 0.4800000000000004t-0.2079999999999984 0.4800000000000004z' })
                )
            );
        }
    }]);

    return FaCode;
}(React.Component);

exports.default = FaCode;
module.exports = exports['default'];