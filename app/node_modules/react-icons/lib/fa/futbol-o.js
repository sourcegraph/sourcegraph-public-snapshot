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

var FaFutbolO = function (_React$Component) {
    _inherits(FaFutbolO, _React$Component);

    function FaFutbolO() {
        _classCallCheck(this, FaFutbolO);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaFutbolO).apply(this, arguments));
    }

    _createClass(FaFutbolO, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm13.594285714285714 18.214285714285715l6.405714285714286-4.642857142857144 6.405714285714286 4.642857142857144-2.4342857142857106 7.5h-7.9228571428571435z m6.405714285714286-18.214285714285715q4.062857142857144 0 7.767142857142858 1.5857142857142859t6.385714285714286 4.2614285714285725 4.261428571428574 6.385714285714284 1.5857142857142819 7.764285714285716-1.585714285714289 7.765714285714285-4.262857142857143 6.385714285714286-6.385714285714286 4.261428571428574-7.765714285714282 1.5857142857142819-7.767142857142858-1.585714285714289-6.385714285714285-4.261428571428574-4.261428571428572-6.385714285714286-1.5857142857142854-7.762857142857136 1.5857142857142859-7.7714285714285705 4.262857142857143-6.385714285714285 6.385714285714286-4.261428571428572 7.765714285714285-1.5814285714285727z m13.817142857142855 30.134285714285717q3.325714285714291-4.531428571428574 3.325714285714291-10.134285714285717v-0.0671428571428585l-2.277142857142856 1.985714285714284-5.357142857142858-5 1.4057142857142857-7.208571428571428 2.991428571428571 0.2671428571428578q-3.3485714285714288-4.6000000000000005-8.682857142857145-6.294285714285714l1.1828571428571415 2.767142857142858-6.405714285714286 3.5500000000000007-6.405714285714286-3.5514285714285707 1.1828571428571433-2.7671428571428573q-5.3342857142857145 1.698571428571428-8.682857142857143 6.2957142857142845l3.0142857142857142-0.2671428571428578 1.3828571428571426 7.209999999999999-5.357142857142858 5-2.2771428571428562-1.985714285714284v0.06571428571428584q0 5.602857142857143 3.325714285714286 10.134285714285717l0.6714285714285717-2.9471428571428575 7.275714285714286 0.894285714285715 3.104285714285714 6.651428571428568-2.59 1.5399999999999991q2.6099999999999994 0.8700000000000045 5.355714285714285 0.8700000000000045t5.357142857142858-0.8714285714285737l-2.59-1.5399999999999991 3.102857142857143-6.651428571428571 7.275714285714283-0.8928571428571423z' })
                )
            );
        }
    }]);

    return FaFutbolO;
}(React.Component);

exports.default = FaFutbolO;
module.exports = exports['default'];