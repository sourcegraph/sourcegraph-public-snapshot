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

var MdFullscreenExit = function (_React$Component) {
    _inherits(MdFullscreenExit, _React$Component);

    function MdFullscreenExit() {
        _classCallCheck(this, MdFullscreenExit);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdFullscreenExit).apply(this, arguments));
    }

    _createClass(MdFullscreenExit, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm26.64 13.360000000000001h5v3.283333333333333h-8.283333333333331v-8.283333333333333h3.2833333333333314v5z m-3.280000000000001 18.28v-8.283333333333331h8.283333333333331v3.2833333333333314h-5v5h-3.2833333333333314z m-10-18.28v-5h3.2833333333333314v8.283333333333331h-8.283333333333333v-3.283333333333333h5z m-5 13.280000000000001v-3.2833333333333314h8.283333333333331v8.283333333333331h-3.283333333333333v-5h-5z' })
                )
            );
        }
    }]);

    return MdFullscreenExit;
}(React.Component);

exports.default = MdFullscreenExit;
module.exports = exports['default'];