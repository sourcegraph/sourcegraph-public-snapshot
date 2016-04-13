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

var FaCaretSquareOLeft = function (_React$Component) {
    _inherits(FaCaretSquareOLeft, _React$Component);

    function FaCaretSquareOLeft() {
        _classCallCheck(this, FaCaretSquareOLeft);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaCaretSquareOLeft).apply(this, arguments));
    }

    _createClass(FaCaretSquareOLeft, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm25.714285714285715 12.857142857142858v14.285714285714288q0 0.5799999999999983-0.4242857142857126 1.0042857142857144t-1.004285714285718 0.4242857142857126q-0.4471428571428575 0-0.8257142857142838-0.2671428571428578l-10-7.142857142857142q-0.6028571428571432-0.42428571428571615-0.6028571428571432-1.1614285714285728t0.6028571428571432-1.1571428571428584l10-7.142857142857142q0.38000000000000256-0.27142857142857046 0.8257142857142838-0.27142857142857046 0.5799999999999983 0 1.0042857142857144 0.4242857142857144t0.42428571428571615 1.0042857142857144z m5.714285714285715 17.857142857142858v-21.42857142857143q0-0.28999999999999915-0.21142857142856997-0.5028571428571436t-0.5028571428571453-0.21142857142856997h-21.42857142857143q-0.28999999999999915 0-0.5028571428571436 0.21142857142857174t-0.21142857142856997 0.5028571428571436v21.42857142857143q0 0.28999999999999915 0.21142857142857174 0.5028571428571418t0.5028571428571436 0.21142857142857352h21.42857142857143q0.28999999999999915 0 0.5028571428571418-0.21142857142856997t0.21142857142857352-0.5028571428571453z m5.714285714285715-21.42857142857143v21.42857142857143q0 2.6571428571428584-1.8857142857142861 4.5428571428571445t-4.5428571428571445 1.8857142857142861h-21.42857142857143q-2.6571428571428575 0-4.542857142857144-1.8857142857142861t-1.885714285714284-4.5428571428571445v-21.42857142857143q0-2.6571428571428575 1.8857142857142857-4.542857142857144t4.542857142857144-1.885714285714284h21.42857142857143q2.6571428571428584 0 4.5428571428571445 1.8857142857142857t1.8857142857142861 4.542857142857144z' })
                )
            );
        }
    }]);

    return FaCaretSquareOLeft;
}(React.Component);

exports.default = FaCaretSquareOLeft;
module.exports = exports['default'];