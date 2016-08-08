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

var FaLocationArrow = function (_React$Component) {
    _inherits(FaLocationArrow, _React$Component);

    function FaLocationArrow() {
        _classCallCheck(this, FaLocationArrow);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaLocationArrow).apply(this, arguments));
    }

    _createClass(FaLocationArrow, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm35.55714285714286 7.790000000000001l-14.285714285714285 28.571428571428577q-0.3771428571428572 0.7814285714285703-1.2714285714285722 0.7814285714285703-0.1114285714285721 0-0.33428571428571274-0.04285714285714448-0.4914285714285711-0.11428571428571388-0.7928571428571445-0.5042857142857144t-0.3014285714285698-0.8814285714285717v-12.857142857142858h-12.857142857142858q-0.491428571428572 0-0.8814285714285726-0.3000000000000007t-0.5028571428571427-0.7942857142857136 0.09142857142857164-0.937142857142856 0.6471428571428568-0.6714285714285708l28.571428571428573-14.285714285714286q0.28857142857143003-0.1542857142857157 0.6457142857142841-0.1542857142857157 0.6028571428571396 0 1.0042857142857144 0.4242857142857144 0.33428571428571274 0.31428571428571406 0.41428571428571104 0.7714285714285714t-0.14714285714286035 0.8799999999999999z' })
                )
            );
        }
    }]);

    return FaLocationArrow;
}(React.Component);

exports.default = FaLocationArrow;
module.exports = exports['default'];