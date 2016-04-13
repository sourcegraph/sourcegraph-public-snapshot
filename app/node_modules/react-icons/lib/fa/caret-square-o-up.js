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

var FaCaretSquareOUp = function (_React$Component) {
    _inherits(FaCaretSquareOUp, _React$Component);

    function FaCaretSquareOUp() {
        _classCallCheck(this, FaCaretSquareOUp);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaCaretSquareOUp).apply(this, arguments));
    }

    _createClass(FaCaretSquareOUp, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm28.414285714285715 24.93285714285714q-0.37714285714285367 0.7814285714285738-1.2714285714285722 0.7814285714285738h-14.285714285714285q-0.8928571428571423 0-1.2714285714285722-0.7814285714285703-0.40285714285714214-0.7814285714285703 0.10999999999999943-1.4714285714285715l7.142857142857142-10q0.42428571428571615-0.6042857142857159 1.1614285714285728-0.6042857142857159t1.1571428571428584 0.6028571428571432l7.142857142857142 10q0.514285714285716 0.6914285714285704 0.11428571428571388 1.4714285714285715z m3.014285714285716 5.781428571428574v-21.42857142857143q0-0.28999999999999915-0.21142857142856997-0.5028571428571436t-0.5028571428571453-0.21142857142856997h-21.42857142857143q-0.28999999999999915 0-0.5028571428571436 0.21142857142857174t-0.21142857142856997 0.5028571428571436v21.42857142857143q0 0.28999999999999915 0.21142857142857174 0.5028571428571418t0.5028571428571436 0.21142857142857352h21.42857142857143q0.28999999999999915 0 0.5028571428571418-0.21142857142856997t0.21142857142857352-0.5028571428571453z m5.714285714285715-21.42857142857143v21.42857142857143q0 2.6571428571428584-1.8857142857142861 4.5428571428571445t-4.5428571428571445 1.8857142857142861h-21.42857142857143q-2.6571428571428575 0-4.542857142857144-1.8857142857142861t-1.885714285714284-4.5428571428571445v-21.42857142857143q0-2.6571428571428575 1.8857142857142857-4.542857142857144t4.542857142857144-1.885714285714284h21.42857142857143q2.6571428571428584 0 4.5428571428571445 1.8857142857142857t1.8857142857142861 4.542857142857144z' })
                )
            );
        }
    }]);

    return FaCaretSquareOUp;
}(React.Component);

exports.default = FaCaretSquareOUp;
module.exports = exports['default'];