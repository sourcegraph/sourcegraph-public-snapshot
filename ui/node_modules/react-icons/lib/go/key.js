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

var GoKey = function (_React$Component) {
    _inherits(GoKey, _React$Component);

    function GoKey() {
        _classCallCheck(this, GoKey);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(GoKey).apply(this, arguments));
    }

    _createClass(GoKey, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm27.534999999999997 2.4962500000000003c-5.524999999999999 0-10 4.475-10 10 0 0.7662499999999994 0.08749999999999858 1.5075000000000003 0.25 2.2225l-15.284999999999997 15.28125v2.5l2.5 2.5h5l2.5-2.5v-2.5h2.5v-2.5h2.5v-2.5h5l2.7662499999999994-2.7650000000000006c0.7300000000000004 0.16750000000000043 1.4875000000000007 0.2575000000000003 2.2699999999999996 0.2575000000000003 5.522499999999997 0 9.999999999999996-4.475000000000001 9.999999999999996-10s-4.481250000000003-9.99625-10-9.99625z m-10.034999999999997 17.50375l-12.5 12.5v-2.5l12.5-12.5v2.5z m12.5-7.5c-1.3787500000000001 0-2.5-1.1212499999999999-2.5-2.5s1.1212499999999999-2.5 2.5-2.5 2.5 1.1212499999999999 2.5 2.5-1.1212499999999999 2.5-2.5 2.5z' })
                )
            );
        }
    }]);

    return GoKey;
}(React.Component);

exports.default = GoKey;
module.exports = exports['default'];