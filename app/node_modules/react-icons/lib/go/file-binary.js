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

var GoFileBinary = function (_React$Component) {
    _inherits(GoFileBinary, _React$Component);

    function GoFileBinary() {
        _classCallCheck(this, GoFileBinary);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(GoFileBinary).apply(this, arguments));
    }

    _createClass(GoFileBinary, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm5 37.5v-35h22.5l7.5 7.5v27.5h-30z m27.5-25l-7.5-7.5h-17.5v30h25v-22.5z m-15 7.5h-7.5v-10h7.5v10z m-2.5-7.5h-2.5v5h2.5v-5z m0 17.5h2.5v2.5h-7.5v-2.5h2.5v-5h-2.5v-2.5h5v7.5z m10-12.5h2.5v2.5h-7.5v-2.5h2.5v-5h-2.5v-2.5h5v7.5z m2.5 15h-7.5v-10h7.5v10z m-2.5-7.5h-2.5v5h2.5v-5z' })
                )
            );
        }
    }]);

    return GoFileBinary;
}(React.Component);

exports.default = GoFileBinary;
module.exports = exports['default'];