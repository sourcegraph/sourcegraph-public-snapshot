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

var MdPermMedia = function (_React$Component) {
    _inherits(MdPermMedia, _React$Component);

    function MdPermMedia() {
        _classCallCheck(this, MdPermMedia);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdPermMedia).apply(this, arguments));
    }

    _createClass(MdPermMedia, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm11.64 25h23.36l-5.859999999999999-7.5-4.140000000000001 5-5.859999999999999-7.5z m25-18.36q1.3283333333333331 0 2.3433333333333337 1.0166666666666666t1.0166666666666657 2.3400000000000007v16.638333333333335q0 1.3283333333333331-1.0166666666666657 2.3433333333333337t-2.3433333333333337 1.0133333333333319h-26.64q-1.3283333333333331 0-2.3433333333333337-1.0166666666666657t-1.0166666666666666-2.341666666666665l0.08000000000000007-20q0-1.3283333333333331 0.9766666666666666-2.3049999999999997t2.3033333333333337-0.9716666666666738h10l3.361666666666668 3.283333333333333h13.283333333333335z m-33.28 3.3599999999999994v23.36h30v3.2833333333333314h-30q-1.3283333333333331 0-2.3433333333333333-0.9783333333333317t-1.0166666666666662-2.306666666666665v-23.358333333333334h3.361666666666667z' })
                )
            );
        }
    }]);

    return MdPermMedia;
}(React.Component);

exports.default = MdPermMedia;
module.exports = exports['default'];