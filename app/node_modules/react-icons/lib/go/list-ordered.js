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

var GoListOrdered = function (_React$Component) {
    _inherits(GoListOrdered, _React$Component);

    function GoListOrdered() {
        _classCallCheck(this, GoListOrdered);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(GoListOrdered).apply(this, arguments));
    }

    _createClass(GoListOrdered, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm17.5 22.5h17.5v-5h-17.5v5z m0 10h17.5v-5h-17.5v5z m0-25v5h17.5v-5h-17.5z m-9.4125 10h3.046249999999999v-10h-1.40625l-3.3200000000000003 0.9000000000000004v1.9499999999999993l1.6800000000000015-0.07499999999999929v7.225z m4.295 8.0475c0-1.40625-0.46875-3.0474999999999994-3.75-3.0474999999999994-1.2874999999999996 0-2.5 0.23499999999999943-3.2424999999999997 0.625l0.037499999999999645 2.5787499999999994c0.82125-0.39124999999999943 1.6425-0.5874999999999986 2.6187500000000004-0.5874999999999986s1.25 0.4312499999999986 1.25 1.0949999999999989c0 1.0150000000000006-1.1724999999999994 2.2650000000000006-4.2975 4.375v1.9525000000000006h7.500000000000001v-2.6174999999999997l-3.556250000000002 0.07874999999999943c1.9124999999999996-1.1712500000000006 3.4000000000000004-2.5775000000000006 3.4000000000000004-4.412500000000001l0.037499999999999645-0.03750000000000142z' })
                )
            );
        }
    }]);

    return GoListOrdered;
}(React.Component);

exports.default = GoListOrdered;
module.exports = exports['default'];