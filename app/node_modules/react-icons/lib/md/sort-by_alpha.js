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

var MdSortByAlpha = function (_React$Component) {
    _inherits(MdSortByAlpha, _React$Component);

    function MdSortByAlpha() {
        _classCallCheck(this, MdSortByAlpha);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdSortByAlpha).apply(this, arguments));
    }

    _createClass(MdSortByAlpha, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm26.25 26.875h10.156666666666666v2.6566666666666663h-14.216666666666669v-2.111666666666668l9.841666666666669-14.296666666666667h-9.766666666666666v-2.6566666666666645h13.828333333333333v2.1116666666666664z m-17.96666666666667-4.141666666666666h6.483333333333334l-3.283333333333333-8.666666666666668z m1.8733333333333348-12.266666666666667h2.7333333333333343l7.5 19.066666666666663h-3.0450000000000017l-1.5633333333333326-4.066666666666666h-8.514999999999999l-1.5633333333333335 4.066666666666666h-3.0466666666666673z m6.953333333333333 21.799999999999997h7.733333333333334l-3.905000000000001 3.905000000000001z m7.811666666666667-24.533333333333335h-7.888333333333332l3.905000000000001-3.9050000000000002z' })
                )
            );
        }
    }]);

    return MdSortByAlpha;
}(React.Component);

exports.default = MdSortByAlpha;
module.exports = exports['default'];