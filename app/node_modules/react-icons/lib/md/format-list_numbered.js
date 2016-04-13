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

var MdFormatListNumbered = function (_React$Component) {
    _inherits(MdFormatListNumbered, _React$Component);

    function MdFormatListNumbered() {
        _classCallCheck(this, MdFormatListNumbered);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdFormatListNumbered).apply(this, arguments));
    }

    _createClass(MdFormatListNumbered, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm11.64 21.64v-3.2833333333333314h23.36v3.2833333333333314h-23.36z m0 10v-3.2833333333333314h23.36v3.2833333333333314h-23.36z m0-23.28h23.36v3.283333333333333h-23.36v-3.283333333333333z m-8.280000000000001 10v-1.7166666666666686h5v1.5616666666666674l-3.046666666666667 3.4383333333333326h3.046666666666667v1.7166666666666686h-5v-1.5616666666666674l2.966666666666667-3.4400000000000013h-2.966666666666667z m1.6400000000000006-5v-5h-1.6400000000000001v-1.7166666666666668h3.2833333333333328v6.716666666666667h-1.6433333333333326z m-1.6400000000000001 15v-1.7166666666666686h5v6.716666666666669h-5v-1.7166666666666686h3.283333333333333v-0.783333333333335h-1.6433333333333326v-1.716666666666665h1.6416666666666666v-0.783333333333335h-3.283333333333333z' })
                )
            );
        }
    }]);

    return MdFormatListNumbered;
}(React.Component);

exports.default = MdFormatListNumbered;
module.exports = exports['default'];