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

var FaObjectGroup = function (_React$Component) {
    _inherits(FaObjectGroup, _React$Component);

    function FaObjectGroup() {
        _classCallCheck(this, FaObjectGroup);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaObjectGroup).apply(this, arguments));
    }

    _createClass(FaObjectGroup, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm40 10h-2.5v20h2.5v7.5h-7.5v-2.5h-25v2.5h-7.5v-7.5h2.5v-20h-2.5v-7.5h7.5v2.5h25v-2.5h7.5v7.5z m-5-5v2.5h2.5v-2.5h-2.5z m-32.5 0v2.5h2.5v-2.5h-2.5z m2.5 30v-2.5h-2.5v2.5h2.5z m27.5-2.5v-2.5h2.5v-20h-2.5v-2.5h-25v2.5h-2.5v20h2.5v2.5h25z m5 2.5v-2.5h-2.5v2.5h2.5z m-12.5-20h7.5v15h-17.5v-5h-7.5v-15h17.5v5z m-15 7.5h12.5v-10h-12.5v10z m20 5v-10h-5v7.5h-7.5v2.5h12.5z' })
                )
            );
        }
    }]);

    return FaObjectGroup;
}(React.Component);

exports.default = FaObjectGroup;
module.exports = exports['default'];