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

var FaCaretUp = function (_React$Component) {
    _inherits(FaCaretUp, _React$Component);

    function FaCaretUp() {
        _classCallCheck(this, FaCaretUp);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaCaretUp).apply(this, arguments));
    }

    _createClass(FaCaretUp, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm31.42857142857143 27.142857142857142q0 0.5800000000000018-0.4242857142857126 1.0042857142857144t-1.004285714285718 0.42428571428571615h-20q-0.5800000000000001 0-1.0042857142857144-0.4242857142857126t-0.4242857142857144-1.004285714285718 0.4242857142857144-1.0042857142857144l10-10q0.42428571428571615-0.4242857142857126 1.0042857142857144-0.4242857142857126t1.0042857142857144 0.4242857142857126l10 10q0.42428571428571615 0.42428571428571615 0.42428571428571615 1.0042857142857144z' })
                )
            );
        }
    }]);

    return FaCaretUp;
}(React.Component);

exports.default = FaCaretUp;
module.exports = exports['default'];