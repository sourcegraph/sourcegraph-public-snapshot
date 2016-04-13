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

var MdLink = function (_React$Component) {
    _inherits(MdLink, _React$Component);

    function MdLink() {
        _classCallCheck(this, MdLink);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdLink).apply(this, arguments));
    }

    _createClass(MdLink, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm28.36 11.64q3.4383333333333326 0 5.859999999999999 2.461666666666666t2.421666666666667 5.898333333333333-2.421666666666667 5.899999999999999-5.859999999999999 2.460000000000001h-6.716666666666669v-3.203333333333333h6.716666666666669q2.1099999999999994 0 3.633333333333333-1.5233333333333334t1.5233333333333334-3.633333333333333-1.5233333333333334-3.633333333333333-3.633333333333333-1.5233333333333334h-6.716666666666669v-3.203333333333333h6.716666666666669z m-15 10v-3.2833333333333314h13.283333333333331v3.2833333333333314h-13.283333333333333z m-6.876666666666666-1.6400000000000006q0 2.1099999999999994 1.5249999999999995 3.633333333333333t3.633333333333333 1.5233333333333334h6.716666666666669v3.203333333333333h-6.716666666666669q-3.4383333333333344 0-5.86-2.461666666666666t-2.4233333333333316-5.898333333333333 2.4250000000000003-5.899999999999999 5.858333333333333-2.460000000000001h6.716666666666667v3.203333333333333h-6.716666666666669q-2.1099999999999994 0-3.633333333333333 1.5233333333333334t-1.5249999999999995 3.633333333333333z' })
                )
            );
        }
    }]);

    return MdLink;
}(React.Component);

exports.default = MdLink;
module.exports = exports['default'];