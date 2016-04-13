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

var FaPowerOff = function (_React$Component) {
    _inherits(FaPowerOff, _React$Component);

    function FaPowerOff() {
        _classCallCheck(this, FaPowerOff);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaPowerOff).apply(this, arguments));
    }

    _createClass(FaPowerOff, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm37.142857142857146 20q0 3.482857142857142-1.3614285714285685 6.651428571428571t-3.6599999999999966 5.468571428571426-5.46857142857143 3.6599999999999966-6.652857142857151 1.3628571428571519-6.65-1.3628571428571448-5.468571428571429-3.6599999999999966-3.6599999999999993-5.46857142857143-1.3614285714285717-6.651428571428575q0-4.062857142857144 1.7971428571428567-7.657142857142858t5.057142857142857-6.024285714285713q0.9571428571428573-0.7142857142857144 2.128571428571428-0.5571428571428569t1.8657142857142865 1.1142857142857148q0.7142857142857135 0.9371428571428568 0.5471428571428572 2.110000000000001t-1.1057142857142868 1.8857142857142861q-2.185714285714287 1.6500000000000004-3.3814285714285717 4.038571428571428t-1.1942857142857122 5.089999999999998q0 2.321428571428573 0.9042857142857148 4.431428571428572t2.442857142857143 3.650000000000002 3.6514285714285712 2.442857142857143 4.431428571428571 0.9057142857142857 4.431428571428572-0.9028571428571439 3.650000000000002-2.442857142857143 2.445714285714285-3.6514285714285712 0.904285714285713-4.431428571428572q0-2.6999999999999993-1.1942857142857157-5.09t-3.3842857142857135-4.040000000000001q-0.937142857142856-0.7142857142857135-1.1042857142857159-1.8857142857142861t0.5457142857142863-2.111428571428572q0.6914285714285704-0.96 1.8757142857142846-1.1142857142857139t2.120000000000001 0.5571428571428569q3.25714285714286 2.4328571428571433 5.054285714285712 6.0285714285714285t1.8028571428571496 7.654285714285715z m-14.285714285714288-17.142857142857142v14.285714285714285q0 1.1600000000000001-0.8485714285714288 2.008571428571429t-2.008571428571429 0.8485714285714288-2.008571428571429-0.8485714285714288-0.8485714285714288-2.008571428571429v-14.285714285714285q0-1.1600000000000006 0.8485714285714288-2.008571428571429t2.008571428571429-0.8485714285714288 2.008571428571429 0.8485714285714285 0.8485714285714288 2.008571428571429z' })
                )
            );
        }
    }]);

    return FaPowerOff;
}(React.Component);

exports.default = FaPowerOff;
module.exports = exports['default'];