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

var MdPlaylistAdd = function (_React$Component) {
    _inherits(MdPlaylistAdd, _React$Component);

    function MdPlaylistAdd() {
        _classCallCheck(this, MdPlaylistAdd);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdPlaylistAdd).apply(this, arguments));
    }

    _createClass(MdPlaylistAdd, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm3.3600000000000003 26.64v-3.2833333333333314h13.283333333333335v3.2833333333333314h-13.283333333333333z m26.64-3.280000000000001h6.640000000000001v3.2833333333333314h-6.640000000000001v6.716666666666669h-3.3599999999999994v-6.716666666666669h-6.640000000000001v-3.2833333333333314h6.640000000000001v-6.716666666666669h3.3599999999999994v6.716666666666669z m-6.640000000000001-13.36v3.3599999999999994h-20v-3.3599999999999994h20z m0 6.640000000000001v3.3599999999999994h-20v-3.3599999999999994h20z' })
                )
            );
        }
    }]);

    return MdPlaylistAdd;
}(React.Component);

exports.default = MdPlaylistAdd;
module.exports = exports['default'];