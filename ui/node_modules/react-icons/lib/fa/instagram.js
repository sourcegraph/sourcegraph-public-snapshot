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

var FaInstagram = function (_React$Component) {
    _inherits(FaInstagram, _React$Component);

    function FaInstagram() {
        _classCallCheck(this, FaInstagram);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaInstagram).apply(this, arguments));
    }

    _createClass(FaInstagram, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm33.25714285714286 31.82857142857143v-14.46142857142857h-3.0114285714285707q0.4471428571428575 1.4057142857142857 0.4471428571428575 2.924285714285716 0 2.814285714285713-1.428571428571427 5.190000000000001t-3.885714285714286 3.7614285714285707-5.357142857142858 1.3828571428571443q-4.395714285714286 0-7.521428571428572-3.024285714285714t-3.1257142857142863-7.308571428571426q0-1.5171428571428578 0.4471428571428575-2.924285714285716h-3.1471428571428595v14.464285714285715q0 0.5799999999999983 0.3899999999999997 0.9714285714285751t0.9714285714285715 0.39000000000000057h23.861428571428576q0.5571428571428569 0 0.9600000000000009-0.39000000000000057t0.3999999999999986-0.9714285714285715z m-6.337142857142858-11.895714285714284q0-2.767142857142858-2.0199999999999996-4.7214285714285715t-4.877142857142857-1.9542857142857173q-2.8342857142857163 0-4.854285714285716 1.9542857142857137t-2.0199999999999996 4.7214285714285715 2.0199999999999996 4.7214285714285715 4.854285714285716 1.952857142857141q2.8571428571428577 0 4.877142857142857-1.952857142857141t2.0199999999999996-4.7214285714285715z m6.338571428571427-8.035714285714286v-3.6828571428571433q0-0.6257142857142854-0.4457142857142884-1.0828571428571427t-1.0942857142857143-0.4571428571428573h-3.885714285714286q-0.6457142857142841 0-1.0928571428571416 0.4571428571428573t-0.4471428571428575 1.0828571428571427v3.6828571428571433q0 0.6471428571428568 0.4471428571428575 1.0942857142857143t1.0942857142857143 0.4471428571428575h3.885714285714286q0.6457142857142841 0 1.0928571428571416-0.4471428571428575t0.4471428571428575-1.0942857142857143z m3.884285714285717-4.642857142857144v25.49142857142857q0 1.808571428571426-1.2942857142857136 3.1028571428571468t-3.1028571428571468 1.2942857142857136h-25.49142857142857q-1.8085714285714296 0-3.102857142857144-1.2942857142857136t-1.2942857142857132-3.1028571428571468v-25.49142857142857q0-1.8085714285714287 1.294285714285714-3.102857142857143t3.1028571428571423-1.294285714285714h25.49142857142857q1.808571428571426 0 3.1028571428571468 1.294285714285714t1.2942857142857136 3.102857142857143z' })
                )
            );
        }
    }]);

    return FaInstagram;
}(React.Component);

exports.default = FaInstagram;
module.exports = exports['default'];