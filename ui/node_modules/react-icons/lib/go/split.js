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

var GoSplit = function (_React$Component) {
    _inherits(GoSplit, _React$Component);

    function GoSplit() {
        _classCallCheck(this, GoSplit);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(GoSplit).apply(this, arguments));
    }

    _createClass(GoSplit, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm17.5 10l-10-10-7.5 7.5 12.15 11.71875c0.5850000000000009-3.1625000000000014 1.678749999999999-5.3125 5.19375-8.985l0.19374999999999787-0.23375000000000057z m5-10l5.195 5.194999999999999-7.695 7.694999999999999c-3.8674999999999997 3.8675000000000015-5 6.328750000000001-5 12.070000000000002v15h10v-15c0-2.03125 0.7424999999999997-3.6724999999999994 2.0700000000000003-5l7.695-7.695 5.195 5.195v-17.5h-17.5z' })
                )
            );
        }
    }]);

    return GoSplit;
}(React.Component);

exports.default = GoSplit;
module.exports = exports['default'];