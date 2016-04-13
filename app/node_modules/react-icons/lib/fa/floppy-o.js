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

var FaFloppyO = function (_React$Component) {
    _inherits(FaFloppyO, _React$Component);

    function FaFloppyO() {
        _classCallCheck(this, FaFloppyO);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaFloppyO).apply(this, arguments));
    }

    _createClass(FaFloppyO, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm11.428571428571429 34.285714285714285h17.14285714285714v-8.57142857142857h-17.14285714285714v8.57142857142857z m20 0h2.857142857142854v-20q0-0.31428571428571317-0.2228571428571442-0.8599999999999977t-0.4471428571428575-0.7714285714285722l-6.271428571428572-6.271428571428571q-0.2228571428571442-0.2228571428571433-0.7571428571428562-0.4471428571428575t-0.8728571428571392-0.22142857142857153v9.285714285714285q0 0.8914285714285715-0.6257142857142846 1.514285714285716t-1.5171428571428578 0.6285714285714263h-12.857142857142858q-0.8928571428571423 0-1.5171428571428578-0.6285714285714299t-0.6257142857142863-1.5142857142857125v-9.285714285714285h-2.8571428571428568v28.57142857142857h2.8571428571428568v-9.285714285714285q0-0.894285714285715 0.6257142857142863-1.518571428571427t1.5171428571428578-0.6242857142857154h18.571428571428573q0.8928571428571423 0 1.5171428571428578 0.6242857142857154t0.6257142857142846 1.518571428571427v9.285714285714285z m-8.571428571428573-20.714285714285715v-7.1428571428571415q0-0.29000000000000004-0.21142857142856997-0.5028571428571427t-0.5028571428571453-0.21142857142857086h-4.285714285714285q-0.28999999999999915 0-0.5028571428571418 0.21142857142857174t-0.21142857142857352 0.5028571428571427v7.142857142857144q0 0.28999999999999915 0.21142857142856997 0.5028571428571436t0.5028571428571453 0.21142857142856997h4.285714285714285q0.28999999999999915 0 0.5028571428571418-0.21142857142857174t0.21142857142857352-0.5028571428571436z m14.285714285714288 0.7142857142857153v20.714285714285715q0 0.8928571428571459-0.6257142857142881 1.5171428571428578t-1.5171428571428578 0.6257142857142881h-30q-0.8928571428571432 0-1.5171428571428573-0.6257142857142881t-0.6257142857142854-1.5171428571428578v-30q0-0.8928571428571432 0.6257142857142859-1.5171428571428573t1.517142857142857-0.6257142857142854h20.714285714285715q0.8928571428571423 0 1.9642857142857153 0.44714285714285706t1.6971428571428575 1.0714285714285712l6.25 6.249999999999999q0.6257142857142881 0.6257142857142863 1.0714285714285694 1.6971428571428575t0.4457142857142884 1.9628571428571444z' })
                )
            );
        }
    }]);

    return FaFloppyO;
}(React.Component);

exports.default = FaFloppyO;
module.exports = exports['default'];