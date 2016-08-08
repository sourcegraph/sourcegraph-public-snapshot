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

var FaCaretSquareORight = function (_React$Component) {
    _inherits(FaCaretSquareORight, _React$Component);

    function FaCaretSquareORight() {
        _classCallCheck(this, FaCaretSquareORight);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaCaretSquareORight).apply(this, arguments));
    }

    _createClass(FaCaretSquareORight, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm27.142857142857142 20q0 0.7371428571428567-0.6028571428571432 1.1600000000000001l-10 7.142857142857142q-0.6914285714285722 0.5142857142857125-1.4714285714285715 0.1114285714285721-0.7828571428571411-0.38000000000000256-0.7828571428571411-1.2714285714285722v-14.285714285714285q0-0.894285714285715 0.781428571428572-1.274285714285714 0.781428571428572-0.40000000000000036 1.4714285714285715 0.1114285714285721l10 7.142857142857142q0.6042857142857159 0.4242857142857126 0.6042857142857159 1.1600000000000001z m4.285714285714285 10.714285714285715v-21.42857142857143q0-0.31428571428571495-0.1999999999999993-0.5142857142857142t-0.5142857142857125-0.1999999999999993h-21.42857142857143q-0.31428571428571495 0-0.5142857142857142 0.1999999999999993t-0.1999999999999993 0.514285714285716v21.42857142857143q0 0.31428571428571317 0.1999999999999993 0.5142857142857125t0.5142857142857142 0.1999999999999993h21.42857142857143q0.31428571428571317 0 0.5142857142857125-0.1999999999999993t0.1999999999999993-0.5142857142857125z m5.714285714285712-21.42857142857143v21.42857142857143q0 2.6571428571428584-1.8857142857142861 4.5428571428571445t-4.542857142857137 1.8857142857142861h-21.42857142857143q-2.6571428571428575 0-4.542857142857144-1.8857142857142861t-1.885714285714284-4.5428571428571445v-21.42857142857143q0-2.6571428571428575 1.8857142857142857-4.542857142857144t4.542857142857144-1.885714285714284h21.42857142857143q2.6571428571428584 0 4.5428571428571445 1.8857142857142857t1.8857142857142861 4.542857142857144z' })
                )
            );
        }
    }]);

    return FaCaretSquareORight;
}(React.Component);

exports.default = FaCaretSquareORight;
module.exports = exports['default'];