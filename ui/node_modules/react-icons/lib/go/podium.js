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

var GoPodium = function (_React$Component) {
    _inherits(GoPodium, _React$Component);

    function GoPodium() {
        _classCallCheck(this, GoPodium);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(GoPodium).apply(this, arguments));
    }

    _createClass(GoPodium, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm17.5 0c-1.25 0-2.5 1.25-2.5 2.5v2.5h-2.5l-7.5 7.5v5h7.5l2.5 15-5 2.5v2.5h20v-2.5l-5-2.5 2.5-15h7.5v-5l-7.5-7.5h-10v-2.5h1.25s1.25-0.6625000000000001 1.25-1.25-0.625-1.25-1.25-1.25h-1.25z m0 32.5l-2.0700000000000003-12.5h4.609999999999999l-0.03750000000000142 12.5h-2.5z m-8.75-20l5-5h1.25v2.5h2.5v-2.5h8.75l5 5h-22.5z' })
                )
            );
        }
    }]);

    return GoPodium;
}(React.Component);

exports.default = GoPodium;
module.exports = exports['default'];