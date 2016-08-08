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

var TiRss = function (_React$Component) {
    _inherits(TiRss, _React$Component);

    function TiRss() {
        _classCallCheck(this, TiRss);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiRss).apply(this, arguments));
    }

    _createClass(TiRss, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm10.003333333333334 26.666666666666668c-1.8450000000000006 0-3.34 1.4933333333333323-3.336666666666666 3.333333333333332 0 1.8399999999999999 1.493333333333334 3.3333333333333357 3.336666666666666 3.3333333333333357 1.838333333333333 0 3.333333333333334-1.4916666666666671 3.33-3.333333333333332 0.0033333333333338544-1.8449999999999989-1.4916666666666671-3.3383333333333347-3.33-3.333333333333332z m-0.0033333333333338544-20c-1.8399999999999999-8.881784197001252e-16-3.333333333333334 1.4933333333333323-3.333333333333334 3.333333333333332s1.493333333333334 3.333333333333334 3.333333333333334 3.333333333333334c9.190000000000001 0 16.666666666666668 7.476666666666668 16.666666666666668 16.666666666666664 0 1.8399999999999999 1.4933333333333323 3.3333333333333357 3.333333333333332 3.3333333333333357s3.3333333333333357-1.4933333333333323 3.3333333333333357-3.333333333333332c0-12.866666666666667-10.466666666666669-23.333333333333336-23.333333333333336-23.333333333333336z m0 10c-1.8399999999999999 0-3.333333333333334 1.4933333333333323-3.333333333333334 3.333333333333332s1.493333333333334 3.333333333333332 3.333333333333334 3.333333333333332c3.6750000000000007 0 6.666666666666668 2.9899999999999984 6.666666666666668 6.666666666666668 0 1.8399999999999999 1.4933333333333323 3.3333333333333357 3.333333333333332 3.3333333333333357s3.333333333333332-1.4933333333333323 3.333333333333332-3.333333333333332c0-7.350000000000001-5.983333333333334-13.333333333333332-13.333333333333334-13.333333333333332z' })
                )
            );
        }
    }]);

    return TiRss;
}(React.Component);

exports.default = TiRss;
module.exports = exports['default'];