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

var FaFacebookOfficial = function (_React$Component) {
    _inherits(FaFacebookOfficial, _React$Component);

    function FaFacebookOfficial() {
        _classCallCheck(this, FaFacebookOfficial);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaFacebookOfficial).apply(this, arguments));
    }

    _createClass(FaFacebookOfficial, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm35.245714285714286 2.857142857142857q0.7814285714285703 0 1.3400000000000034 0.5571428571428574t0.5571428571428569 1.3428571428571425v30.490000000000006q0 0.7814285714285703-0.5571428571428569 1.3400000000000034t-1.3400000000000034 0.5600000000000023h-8.728571428571428v-13.28142857142857h4.442857142857143l0.6714285714285708-5.178571428571431h-5.114285714285714v-3.3028571428571425q0-1.25 0.5257142857142867-1.8757142857142863t2.0428571428571445-0.6257142857142863l2.722857142857137-0.025714285714293794v-4.617142857142857q-1.4057142857142857-0.1999999999999993-3.9714285714285715-0.1999999999999993-3.0371428571428574 0-4.857142857142858 1.7857142857142865t-1.8171428571428585 5.042857142857143v3.8171428571428567h-4.462857142857139v5.177142857142858h4.462857142857143v13.280000000000001h-16.405714285714286q-0.7814285714285707 0-1.339999999999999-0.5571428571428569t-0.5571428571428569-1.3400000000000034v-30.49142857142857q0-0.7814285714285716 0.5571428571428569-1.3399999999999999t1.3399999999999999-0.5571428571428574h30.491428571428575z' })
                )
            );
        }
    }]);

    return FaFacebookOfficial;
}(React.Component);

exports.default = FaFacebookOfficial;
module.exports = exports['default'];