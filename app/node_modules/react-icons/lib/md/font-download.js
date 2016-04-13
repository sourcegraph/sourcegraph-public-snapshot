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

var MdFontDownload = function (_React$Component) {
    _inherits(MdFontDownload, _React$Component);

    function MdFontDownload() {
        _classCallCheck(this, MdFontDownload);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdFontDownload).apply(this, arguments));
    }

    _createClass(MdFontDownload, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm26.563333333333336 30.86h3.5166666666666657l-8.51666666666667-21.716666666666665h-3.125l-8.516666666666667 21.716666666666665h3.5166666666666675l1.875-5h9.375z m6.79666666666667-27.5q1.3283333333333331 0 2.3049999999999997 0.9766666666666666t0.9750000000000014 2.3049999999999997v26.71666666666667q0 1.3299999999999983-0.9766666666666666 2.306666666666665t-2.306666666666665 0.9750000000000014h-26.713333333333342q-1.330000000000001 0-2.3066666666666675-0.9766666666666666t-0.9766666666666666-2.306666666666665v-26.713333333333335q0-1.330000000000001 0.9766666666666666-2.3066666666666675t2.3050000000000006-0.9766666666666666h26.71666666666667z m-16.79666666666667 19.14l3.436666666666664-9.216666666666667 3.4400000000000013 9.216666666666667h-6.873333333333335z' })
                )
            );
        }
    }]);

    return MdFontDownload;
}(React.Component);

exports.default = MdFontDownload;
module.exports = exports['default'];