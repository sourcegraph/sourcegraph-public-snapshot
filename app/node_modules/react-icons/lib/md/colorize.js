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

var MdColorize = function (_React$Component) {
    _inherits(MdColorize, _React$Component);

    function MdColorize() {
        _classCallCheck(this, MdColorize);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdColorize).apply(this, arguments));
    }

    _createClass(MdColorize, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm11.563333333333333 31.640000000000004l13.436666666666667-13.440000000000005-3.1999999999999993-3.1999999999999993-13.441666666666668 13.433333333333334z m22.96666666666667-22.266666666666666q1.173333333333332 1.1733333333333338 0 2.3450000000000006l-5.233333333333334 5.236666666666666 3.203333333333333 3.203333333333333-2.3433333333333337 2.3433333333333337-2.3433333333333337-2.3433333333333337-14.921666666666665 14.841666666666661h-7.8916666666666675v-7.890000000000001l14.844999999999999-14.921666666666667-2.344999999999999-2.343333333333332 2.3466666666666676-2.3450000000000006 3.203333333333333 3.205 5.233333333333334-5.233333333333333q0.466666666666665-0.4716666666666667 1.1716666666666633-0.4716666666666667t1.1716666666666669 0.46999999999999975z' })
                )
            );
        }
    }]);

    return MdColorize;
}(React.Component);

exports.default = MdColorize;
module.exports = exports['default'];