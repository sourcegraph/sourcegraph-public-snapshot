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

var MdFormatIndentIncrease = function (_React$Component) {
    _inherits(MdFormatIndentIncrease, _React$Component);

    function MdFormatIndentIncrease() {
        _classCallCheck(this, MdFormatIndentIncrease);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdFormatIndentIncrease).apply(this, arguments));
    }

    _createClass(MdFormatIndentIncrease, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm18.36 21.64v-3.2833333333333314h16.64v3.2833333333333314h-16.64z m0-6.640000000000001v-3.3599999999999994h16.64v3.3599999999999994h-16.64z m-13.36-10h30v3.3599999999999994h-30v-3.3599999999999994z m13.36 23.36v-3.3599999999999994h16.64v3.3599999999999994h-16.64z m-13.36-14.999999999999998l6.640000000000001 6.639999999999999-6.640000000000001 6.640000000000001v-13.283333333333333z m0 21.64v-3.3599999999999994h30v3.3599999999999994h-30z' })
                )
            );
        }
    }]);

    return MdFormatIndentIncrease;
}(React.Component);

exports.default = MdFormatIndentIncrease;
module.exports = exports['default'];