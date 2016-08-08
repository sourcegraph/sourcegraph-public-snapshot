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

var FaHeartbeat = function (_React$Component) {
    _inherits(FaHeartbeat, _React$Component);

    function FaHeartbeat() {
        _classCallCheck(this, FaHeartbeat);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaHeartbeat).apply(this, arguments));
    }

    _createClass(FaHeartbeat, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm28.571428571428573 22.857142857142858h6.80857142857143q-0.11142857142856855 0.13428571428571345-0.2228571428571442 0.23428571428571487t-0.20000000000000284 0.16714285714285637l-0.06857142857143117 0.09142857142857252-13.905714285714286 13.392857142857146q-0.3999999999999986 0.3999999999999986-0.9828571428571422 0.3999999999999986t-0.9828571428571422-0.3999999999999986l-13.928571428571429-13.438571428571429q-0.11142857142857121-0.04285714285714448-0.46857142857142886-0.4471428571428575h8.23714285714286q0.4914285714285711 0 0.8814285714285717-0.3000000000000007t0.5028571428571436-0.7714285714285722l1.5614285714285714-6.271428571428572 4.242857142857142 14.885714285714286q0.13285714285714434 0.4485714285714302 0.5114285714285707 0.7385714285714293t0.8714285714285737 0.28999999999999915q0.46857142857142975 0 0.8485714285714288-0.28999999999999915t0.514285714285716-0.7371428571428567l3.2571428571428562-10.824285714285715 1.25 2.5q0.3999999999999986 0.7814285714285703 1.2714285714285722 0.7814285714285703z m11.428571428571427-9.554285714285713q0 3.2371428571428584-2.299999999999997 6.697142857142856h-8.235714285714288l-2.4771428571428586-4.9328571428571415q-0.17714285714285793-0.3800000000000008-0.5671428571428585-0.6028571428571432t-0.8142857142857132-0.17857142857142883q-1.0042857142857144 0.1114285714285721-1.25 1.0285714285714285l-2.879999999999999 9.597142857142858-4.375714285714285-15.314285714285715q-0.13571428571428612-0.4457142857142866-0.5257142857142867-0.7357142857142858t-0.8814285714285717-0.2900000000000009-0.8714285714285719 0.3000000000000007-0.4900000000000002 0.7714285714285722l-2.588571428571427 10.357142857142856h-9.442857142857143q-2.3000000000000007-3.460000000000001-2.3000000000000007-6.6971428571428575 0-4.91 2.835714285714286-7.678571428571429t7.835714285714285-2.7671428571428565q1.3828571428571426 0 2.822857142857142 0.48t2.678571428571427 1.2942857142857145 2.1328571428571443 1.5285714285714285 1.6942857142857157 1.517142857142857q0.8042857142857152-0.8028571428571434 1.6999999999999993-1.517142857142857t2.12857142857143-1.5285714285714285 2.6799999999999997-1.2942857142857145 2.8242857142857147-0.48q5 0 7.834285714285713 2.7671428571428573t2.8328571428571436 7.678571428571428z' })
                )
            );
        }
    }]);

    return FaHeartbeat;
}(React.Component);

exports.default = FaHeartbeat;
module.exports = exports['default'];