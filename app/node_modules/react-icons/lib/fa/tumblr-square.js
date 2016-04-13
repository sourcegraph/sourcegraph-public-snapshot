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

var FaTumblrSquare = function (_React$Component) {
    _inherits(FaTumblrSquare, _React$Component);

    function FaTumblrSquare() {
        _classCallCheck(this, FaTumblrSquare);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaTumblrSquare).apply(this, arguments));
    }

    _createClass(FaTumblrSquare, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm28.214285714285715 32.61142857142857l-1.3857142857142861-4.085714285714285q-0.9814285714285731 0.49285714285714377-2.297142857142859 0.49285714285714377-0.8028571428571425 0.022857142857141355-1.3857142857142861-0.23428571428571487t-0.8571428571428577-0.7042857142857137-0.39142857142857324-0.904285714285713-0.1114285714285721-0.9714285714285715v-8.885714285714286h5.737142857142857v-4.328571428571429h-5.714285714285715v-7.275714285714284h-4.1971428571428575q-0.17857142857142705 0-0.1999999999999993 0.2228571428571433-0.11428571428571388 0.9828571428571431-0.3914285714285697 1.942857142857143t-0.8714285714285701 2.1185714285714283-1.717142857142857 2.1185714285714283-2.6457142857142824 1.5185714285714287v3.6828571428571415h2.9000000000000004v9.32857142857143q0 1.274285714285714 0.4814285714285713 2.5685714285714276t1.4514285714285702 2.4785714285714278 2.6999999999999993 1.9085714285714275 3.942857142857143 0.682857142857145q1.53857142857143-0.022857142857141355 3.0457142857142863-0.5571428571428569t1.9085714285714275-1.1171428571428592z m8.92857142857143-23.325714285714284v21.42857142857143q0 2.6571428571428584-1.8857142857142861 4.5428571428571445t-4.5428571428571445 1.8857142857142861h-21.42857142857143q-2.6571428571428575 0-4.542857142857144-1.8857142857142861t-1.885714285714284-4.5428571428571445v-21.42857142857143q0-2.6571428571428575 1.8857142857142857-4.542857142857144t4.542857142857144-1.885714285714284h21.42857142857143q2.6571428571428584 0 4.5428571428571445 1.8857142857142857t1.8857142857142861 4.542857142857144z' })
                )
            );
        }
    }]);

    return FaTumblrSquare;
}(React.Component);

exports.default = FaTumblrSquare;
module.exports = exports['default'];