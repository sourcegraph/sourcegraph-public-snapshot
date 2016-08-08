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

var MdDeveloperMode = function (_React$Component) {
    _inherits(MdDeveloperMode, _React$Component);

    function MdDeveloperMode() {
        _classCallCheck(this, MdDeveloperMode);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdDeveloperMode).apply(this, arguments));
    }

    _createClass(MdDeveloperMode, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm28.36 31.640000000000004v-3.283333333333335h3.2833333333333314v6.643333333333331q0 1.326666666666668-0.9783333333333317 2.3416666666666686t-2.306666666666665 1.0166666666666657h-16.715000000000003q-1.3299999999999983 0-2.306666666666665-1.0166666666666657t-0.9766666666666648-2.3416666666666686v-6.641666666666666h3.283333333333333v3.2833333333333314h16.716666666666665z m-11.719999999999999-6.326666666666664l-2.3433333333333337 2.3433333333333337-7.656666666666666-7.656666666666673 7.656666666666666-7.656666666666666 2.3433333333333337 2.3433333333333337-5.2333333333333325 5.313333333333333z m9.063333333333333 2.3433333333333337l-2.3433333333333337-2.3433333333333337 5.233333333333334-5.31333333333334-5.233333333333334-5.313333333333333 2.3433333333333337-2.3433333333333337 7.656666666666666 7.656666666666666z m-14.063333333333333-19.296666666666674v3.283333333333335h-3.283333333333333v-6.643333333333334q0-1.326666666666667 0.9783333333333335-2.341666666666667t2.3066666666666666-1.0166666666666666l16.71666666666667 0.08000000000000007q1.3299999999999983 0 2.306666666666665 0.9766666666666666t0.975000000000005 2.3016666666666667v6.640000000000001h-3.283333333333335v-3.283333333333333h-16.715000000000003z' })
                )
            );
        }
    }]);

    return MdDeveloperMode;
}(React.Component);

exports.default = MdDeveloperMode;
module.exports = exports['default'];