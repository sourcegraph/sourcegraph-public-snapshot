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

var MdNowWallpaper = function (_React$Component) {
    _inherits(MdNowWallpaper, _React$Component);

    function MdNowWallpaper() {
        _classCallCheck(this, MdNowWallpaper);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdNowWallpaper).apply(this, arguments));
    }

    _createClass(MdNowWallpaper, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm6.640000000000001 21.64v11.716666666666669h11.716666666666669v3.2833333333333314h-11.716666666666667q-1.3283333333333331 0-2.3049999999999997-0.9766666666666666t-0.9750000000000001-2.3049999999999997v-11.716666666666669h3.2833333333333337z m26.72 11.719999999999999v-11.716666666666669h3.2833333333333314v11.716666666666669q0 1.3283333333333331-0.9783333333333317 2.3049999999999997t-2.306666666666665 0.9750000000000014h-11.716666666666669v-3.2833333333333314h11.716666666666669z m0-30q1.3283333333333331 0 2.3049999999999997 0.9766666666666666t0.9750000000000014 2.3049999999999997v11.716666666666669h-3.2833333333333314v-11.716666666666667h-11.716666666666669v-3.283333333333333h11.716666666666669z m-5 10.780000000000001q0 1.0166666666666657-0.7416666666666671 1.7583333333333329t-1.7566666666666677 0.7399999999999984-1.7583333333333329-0.7416666666666671-0.7399999999999984-1.7599999999999998 0.7416666666666671-1.7583333333333329 1.7600000000000016-0.7433333333333341 1.7583333333333329 0.7416666666666671 0.7433333333333323 1.7566666666666677z m-11.719999999999999 7.5l5 6.171666666666667 3.3599999999999994-4.453333333333333 5 6.641666666666666h-20z m-10-15v11.716666666666669h-3.283333333333333v-11.716666666666667q0-1.3283333333333331 0.9783333333333335-2.3049999999999997t2.3066666666666666-0.9750000000000001h11.716666666666667v3.2833333333333337h-11.716666666666667z' })
                )
            );
        }
    }]);

    return MdNowWallpaper;
}(React.Component);

exports.default = MdNowWallpaper;
module.exports = exports['default'];