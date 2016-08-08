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

var GoBook = function (_React$Component) {
    _inherits(GoBook, _React$Component);

    function GoBook() {
        _classCallCheck(this, GoBook);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(GoBook).apply(this, arguments));
    }

    _createClass(GoBook, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm30 22.5h-5c-1.3287499999999994 0-2.5 1.25-2.5 2.5h10c0-1.3287499999999994-1.25-2.5-2.5-2.5z m-2.1499999999999986-16.25c-6.521250000000002 0-8.162500000000001 1.25-9.100000000000001 2.1875-0.9375-0.9375-2.5787499999999994-2.1875-9.1-2.1875s-9.65 1.7974999999999994-9.65 3.0474999999999994v22.89c1.1325-0.625 4.65-1.875 8.36-2.1875 4.4925-0.3500000000000014 9.14125 0.35249999999999915 9.14125 1.25 0 0.625 0.3125 1.2124999999999986 1.2124999999999986 1.25h0.07750000000000057c0.8999999999999986-0.03750000000000142 1.2124999999999986-0.625 1.2124999999999986-1.25 0-0.8975000000000009 4.647500000000001-1.6000000000000014 9.14-1.25 3.6724999999999994 0.2749999999999986 7.225000000000001 1.5625 8.36 2.1875v-22.8875c0-1.25-3.125-3.0475000000000003-9.649999999999999-3.0475000000000003z m-10.350000000000001 22.34375c-1.1724999999999994-0.625-4.025-1.09375-7.5-1.09375s-6.641249999999999 0.46875-7.5 1.0549999999999997v-17.305s2.5-2.3049999999999997 7.5-2.3049999999999997 7.5 1.0549999999999997 7.5 2.3049999999999997v17.34375z m17.5-0.03750000000000142c-0.8599999999999994-0.5874999999999986-4.024999999999999-1.0562499999999986-7.5-1.0562499999999986s-6.328749999999999 0.46875-7.5 1.09375v-17.34375s2.5-2.3049999999999997 7.5-2.3049999999999997 7.5 1.0549999999999997 7.5 2.3049999999999997v17.305z m-5-11.056249999999999h-5c-1.3287499999999994 0-2.5 1.25-2.5 2.5h10c0-1.3287499999999994-1.25-2.5-2.5-2.5z m0-5h-5c-1.3287499999999994 0-2.5 1.25-2.5 2.5h10c0-1.3287499999999994-1.25-2.5-2.5-2.5z m-17.5 5h-5c-1.25 0-2.5 1.1724999999999994-2.5 2.5h10c0-1.25-1.1724999999999994-2.5-2.5-2.5z m0 5h-5c-1.25 0-2.5 1.1724999999999994-2.5 2.5h10c0-1.25-1.1724999999999994-2.5-2.5-2.5z m0-10h-5c-1.25 0-2.5 1.1724999999999994-2.5 2.5h10c0-1.25-1.1724999999999994-2.5-2.5-2.5z' })
                )
            );
        }
    }]);

    return GoBook;
}(React.Component);

exports.default = GoBook;
module.exports = exports['default'];