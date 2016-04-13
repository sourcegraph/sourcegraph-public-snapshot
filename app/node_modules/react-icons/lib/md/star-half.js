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

var MdStarHalf = function (_React$Component) {
    _inherits(MdStarHalf, _React$Component);

    function MdStarHalf() {
        _classCallCheck(this, MdStarHalf);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdStarHalf).apply(this, arguments));
    }

    _createClass(MdStarHalf, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm20 25.703333333333337l6.25 3.75-1.6400000000000006-7.109999999999999 5.546666666666667-4.843333333333334-7.343333333333334-0.625-2.8133333333333326-6.71666666666667v15.545z m16.64-10.313333333333333l-9.063333333333333 7.890000000000001 2.7333333333333343 11.716666666666665-10.310000000000002-6.24666666666667-10.316666666666666 6.25 2.7366666666666664-11.716666666666669-9.063333333333333-7.890000000000001 11.953333333333333-1.0166666666666657 4.6899999999999995-11.014999999999999 4.686666666666667 11.016666666666667z' })
                )
            );
        }
    }]);

    return MdStarHalf;
}(React.Component);

exports.default = MdStarHalf;
module.exports = exports['default'];