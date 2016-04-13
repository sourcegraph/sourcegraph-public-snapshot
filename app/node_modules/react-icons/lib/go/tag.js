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

var GoTag = function (_React$Component) {
    _inherits(GoTag, _React$Component);

    function GoTag() {
        _classCallCheck(this, GoTag);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(GoTag).apply(this, arguments));
    }

    _createClass(GoTag, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm17.5 2.5h-10l-5 5v10l20 20 15-15-20-20z m-12.5 13.75v-7.5l3.75-3.75h7.5l17.5 17.5-11.25 11.25-17.5-17.5z m15-3.75l-7.5 7.5 10 10 7.5-7.5-10-10z m-3.75 7.5l3.75-3.75 6.25 6.25-3.75 3.75-6.25-6.25z m-1.25-8.75c0-2.0700000000000003-1.6799999999999997-3.75-3.75-3.75s-3.75 1.6799999999999997-3.75 3.75 1.6799999999999997 3.75 3.75 3.75 3.75-1.6799999999999997 3.75-3.75z m-3.75 1.25c-0.6899999999999995 0-1.25-0.5600000000000005-1.25-1.25s0.5600000000000005-1.25 1.25-1.25 1.25 0.5600000000000005 1.25 1.25-0.5600000000000005 1.25-1.25 1.25z' })
                )
            );
        }
    }]);

    return GoTag;
}(React.Component);

exports.default = GoTag;
module.exports = exports['default'];