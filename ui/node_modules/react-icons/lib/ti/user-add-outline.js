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

var TiUserAddOutline = function (_React$Component) {
    _inherits(TiUserAddOutline, _React$Component);

    function TiUserAddOutline() {
        _classCallCheck(this, TiUserAddOutline);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiUserAddOutline).apply(this, arguments));
    }

    _createClass(TiUserAddOutline, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm35 23.333333333333336h-10c-0.9216666666666669 0-1.6666666666666679-0.7466666666666661-1.6666666666666679-1.6666666666666679s0.745000000000001-1.6666666666666679 1.6666666666666679-1.6666666666666679h10c0.9216666666666669 0 1.6666666666666643 0.7466666666666661 1.6666666666666643 1.6666666666666679s-0.7449999999999974 1.6666666666666679-1.6666666666666643 1.6666666666666679z m-5 5c-0.9216666666666669 0-1.6666666666666679-0.7466666666666661-1.6666666666666679-1.6666666666666679v-10c0-0.9199999999999999 0.745000000000001-1.666666666666666 1.6666666666666679-1.666666666666666s1.6666666666666679 0.7466666666666661 1.6666666666666679 1.666666666666666v10c0 0.9200000000000017-0.745000000000001 1.6666666666666679-1.6666666666666679 1.6666666666666679z m-15-18.333333333333336c2.7566666666666677 0 5 2.243333333333334 5 5s-2.2433333333333323 5-5 5-5-2.2433333333333323-5-5 2.243333333333334-5 5-5z m0-3.333333333333334c-4.606666666666667 8.881784197001252e-16-8.333333333333332 3.730000000000002-8.333333333333332 8.333333333333334s3.7266666666666666 8.333333333333336 8.333333333333332 8.333333333333336 8.333333333333336-3.7300000000000004 8.333333333333336-8.333333333333334-3.7266666666666666-8.333333333333334-8.333333333333334-8.333333333333334z m0 21.66666666666667c3.366666666666667 0 5.5 1.283333333333335 6.305 2.408333333333335-1.1383333333333319 0.43333333333333357-3.2833333333333314 0.9250000000000007-6.305 0.9250000000000007-3.3066666666666666 0-5.343333333333334-0.5083333333333329-6.363333333333333-0.9033333333333324 0.7666666666666675-1.1916666666666735 2.963333333333333-2.4300000000000033 6.363333333333333-2.4300000000000033z m0-3.333333333333332c-6.25 0-10 3.333333333333332-10 6.666666666666668 0 1.6666666666666643 3.75 3.3333333333333357 10 3.3333333333333357 5.863333333333333 0 10-1.6666666666666643 10-3.333333333333332s-3.923333333333332-6.666666666666668-10-6.666666666666668z' })
                )
            );
        }
    }]);

    return TiUserAddOutline;
}(React.Component);

exports.default = TiUserAddOutline;
module.exports = exports['default'];