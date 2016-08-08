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

var MdMoveToInbox = function (_React$Component) {
    _inherits(MdMoveToInbox, _React$Component);

    function MdMoveToInbox() {
        _classCallCheck(this, MdMoveToInbox);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdMoveToInbox).apply(this, arguments));
    }

    _createClass(MdMoveToInbox, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm26.64 16.64l-6.640000000000001 6.716666666666665-6.640000000000001-6.716666666666669h3.2833333333333314v-5h6.716666666666669v5h3.2833333333333314z m5 8.36v-16.64h-23.36v16.64h6.716666666666669q0 2.0333333333333314 1.4866666666666681 3.5166666666666657t3.516666666666662 1.4833333333333343 3.5133333333333354-1.4833333333333343 1.4833333333333343-3.5166666666666657h6.641666666666666z m0-20q1.3283333333333331 0 2.3433333333333337 1.0166666666666666t1.0166666666666657 2.3400000000000007v23.28333333333333q0 1.326666666666668-1.0166666666666657 2.3416666666666686t-2.3433333333333337 1.0166666666666657h-23.36q-1.4066666666666663 0-2.3433333333333337-0.9783333333333317t-0.9366666666666674-2.3816666666666677v-23.283333333333335q0-1.4050000000000002 0.9366666666666665-2.383333333333333t2.3433333333333346-0.9716666666666676h23.36z' })
                )
            );
        }
    }]);

    return MdMoveToInbox;
}(React.Component);

exports.default = MdMoveToInbox;
module.exports = exports['default'];