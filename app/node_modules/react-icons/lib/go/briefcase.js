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

var GoBriefcase = function (_React$Component) {
    _inherits(GoBriefcase, _React$Component);

    function GoBriefcase() {
        _classCallCheck(this, GoBriefcase);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(GoBriefcase).apply(this, arguments));
    }

    _createClass(GoBriefcase, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm35 7.5h-10v-2.5787500000000003c0-1.3374999999999995-1.0875000000000021-2.4212499999999997-2.4224999999999994-2.4212499999999997h-5.15625c-1.3374999999999986 0-2.4212500000000006 1.0837500000000002-2.4212500000000006 2.42v2.58h-10c-1.38 0-2.5 1.1212499999999999-2.5 2.5v20c0 1.3787500000000001 1.12 2.5 2.5 2.5h30c1.3774999999999977 0 2.5-1.1212499999999999 2.5-2.5v-20c0-1.3787500000000001-1.1225000000000023-2.5-2.5-2.5z m-17.5-1.875c0-0.34375 0.28125-0.625 0.625-0.625h3.75c0.34375 0 0.625 0.28125 0.625 0.625v1.875h-5v-1.875z m17.5 14.375h-12.5v2.5h-5v-2.5h-12.5v-10h2.5v7.5h25v-7.5h2.5v10z' })
                )
            );
        }
    }]);

    return GoBriefcase;
}(React.Component);

exports.default = GoBriefcase;
module.exports = exports['default'];