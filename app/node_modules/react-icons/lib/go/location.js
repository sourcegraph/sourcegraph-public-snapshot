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

var GoLocation = function (_React$Component) {
    _inherits(GoLocation, _React$Component);

    function GoLocation() {
        _classCallCheck(this, GoLocation);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(GoLocation).apply(this, arguments));
    }

    _createClass(GoLocation, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm20 0c-6.912500000000001 0-12.5 5.5874999999999995-12.5 12.5s6.25 16.25 12.5 27.5c6.25-11.25 12.5-20.5875 12.5-27.5s-5.587499999999999-12.5-12.5-12.5z m0 17.5c-2.7749999999999986 0-5-2.2249999999999996-5-5s2.2250000000000014-5 5-5 5 2.2249999999999996 5 5-2.2250000000000014 5-5 5z' })
                )
            );
        }
    }]);

    return GoLocation;
}(React.Component);

exports.default = GoLocation;
module.exports = exports['default'];