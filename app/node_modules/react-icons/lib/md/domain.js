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

var MdDomain = function (_React$Component) {
    _inherits(MdDomain, _React$Component);

    function MdDomain() {
        _classCallCheck(this, MdDomain);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdDomain).apply(this, arguments));
    }

    _createClass(MdDomain, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm30 25v3.3599999999999994h-3.3599999999999994v-3.3599999999999994h3.3599999999999994z m0-6.640000000000001v3.2833333333333314h-3.3599999999999994v-3.2833333333333314h3.3599999999999994z m3.3599999999999994 13.280000000000001v-16.64h-13.36v3.3599999999999994h3.3599999999999994v3.2833333333333314h-3.3599999999999994v3.356666666666669h3.3599999999999994v3.361666666666668h-3.3599999999999994v3.283333333333335h13.36z m-16.720000000000002-20v-3.283333333333333h-3.283333333333333v3.283333333333333h3.283333333333333z m0 6.719999999999999v-3.3599999999999994h-3.283333333333333v3.3599999999999994h3.283333333333333z m0 6.640000000000001v-3.3599999999999994h-3.283333333333333v3.3599999999999994h3.283333333333333z m0 6.640000000000001v-3.2833333333333314h-3.283333333333333v3.2833333333333314h3.283333333333333z m-6.639999999999997-20v-3.283333333333333h-3.3599999999999994v3.283333333333333h3.3599999999999994z m0 6.719999999999999v-3.3599999999999994h-3.3599999999999994v3.3599999999999994h3.3599999999999994z m0 6.640000000000001v-3.3599999999999994h-3.3599999999999994v3.3599999999999994h3.3599999999999994z m0 6.640000000000001v-3.2833333333333314h-3.3599999999999994v3.2833333333333314h3.3599999999999994z m10-20h16.64v23.36h-33.28333333333333v-30h16.64333333333333v6.640000000000001z' })
                )
            );
        }
    }]);

    return MdDomain;
}(React.Component);

exports.default = MdDomain;
module.exports = exports['default'];