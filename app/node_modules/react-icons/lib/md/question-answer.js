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

var MdQuestionAnswer = function (_React$Component) {
    _inherits(MdQuestionAnswer, _React$Component);

    function MdQuestionAnswer() {
        _classCallCheck(this, MdQuestionAnswer);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdQuestionAnswer).apply(this, arguments));
    }

    _createClass(MdQuestionAnswer, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm28.36 20q0 0.7033333333333331-0.5083333333333329 1.1716666666666669t-1.2100000000000009 0.466666666666665h-16.641666666666666l-6.640000000000001 6.720000000000002v-23.358333333333334q0-0.7050000000000001 0.4666666666666668-1.1733333333333333t1.1733333333333338-0.46666666666666634h21.64q0.7033333333333331 0 1.211666666666666 0.4666666666666668t0.5100000000000016 1.1716666666666669v15z m6.640000000000001-10q0.7033333333333331 0 1.1716666666666669 0.4666666666666668t0.46666666666666856 1.1733333333333338v25l-6.638333333333335-6.640000000000001h-18.36q-0.7033333333333331 0-1.1716666666666669-0.466666666666665t-0.4683333333333337-1.1750000000000007v-3.3583333333333343h21.64v-15h3.3599999999999994z' })
                )
            );
        }
    }]);

    return MdQuestionAnswer;
}(React.Component);

exports.default = MdQuestionAnswer;
module.exports = exports['default'];