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

var TiImage = function (_React$Component) {
    _inherits(TiImage, _React$Component);

    function TiImage() {
        _classCallCheck(this, TiImage);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiImage).apply(this, arguments));
    }

    _createClass(TiImage, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm18.333333333333336 14.166666666666668c0 2.3049999999999997-1.8666666666666671 4.166666666666668-4.166666666666668 4.166666666666668s-4.166666666666668-1.8616666666666681-4.166666666666668-4.166666666666668c0-2.3000000000000007 1.8666666666666671-4.166666666666668 4.166666666666668-4.166666666666668s4.166666666666668 1.8666666666666671 4.166666666666668 4.166666666666668z m8.333333333333332 2.5c-3.333333333333332 0-5 5-7.5 5s-2.5-1.6666666666666679-5.833333333333334-1.6666666666666679-5 6.666666666666668-5 6.666666666666668h23.333333333333336s-1.6666666666666679-10-5-10z m6.666666666666668-11.666666666666668h-26.666666666666668c-1.8383333333333347 0-3.3333333333333344 1.495-3.3333333333333344 3.333333333333334v20c0 1.8383333333333347 1.4949999999999997 3.333333333333332 3.3333333333333335 3.333333333333332h26.666666666666668c1.8383333333333312 0 3.3333333333333357-1.495000000000001 3.3333333333333357-3.333333333333332v-20c0-1.8383333333333347-1.4949999999999974-3.3333333333333357-3.3333333333333357-3.3333333333333357z m0 23.333333333333336h-26.666666666666668v-20h26.666666666666668v20z' })
                )
            );
        }
    }]);

    return TiImage;
}(React.Component);

exports.default = TiImage;
module.exports = exports['default'];