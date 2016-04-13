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

var TiMediaEjectOutline = function (_React$Component) {
    _inherits(TiMediaEjectOutline, _React$Component);

    function TiMediaEjectOutline() {
        _classCallCheck(this, TiMediaEjectOutline);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiMediaEjectOutline).apply(this, arguments));
    }

    _createClass(TiMediaEjectOutline, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm26.666666666666668 35h-13.333333333333334c-2.756666666666666 0-5-2.2433333333333323-5-5s2.243333333333334-5 5-5h13.333333333333334c2.7566666666666677 0 5 2.2433333333333323 5 5s-2.2433333333333323 5-5 5z m-13.333333333333334-6.666666666666668c-0.9166666666666661 0-1.666666666666666 0.7466666666666661-1.666666666666666 1.6666666666666679s0.75 1.6666666666666679 1.666666666666666 1.6666666666666679h13.333333333333334c0.9166666666666679 0 1.6666666666666679-0.7466666666666661 1.6666666666666679-1.6666666666666679s-0.75-1.6666666666666679-1.6666666666666679-1.6666666666666679h-13.333333333333334z m6.666666666666666-16.89l8.273333333333333 8.493333333333336 0.060000000000002274 0.06333333333333258-16.666666666666668 0.006666666666667709 0.13333333333333286-0.14499999999999957 8.2-8.416666666666668z m0-4.776666666666665l-10.721666666666668 11.006666666666668c-0.5833333333333321 0.6050000000000004-0.9449999999999985 1.4216666666666669-0.9449999999999985 2.3266666666666644 0 1.8399999999999999 1.493333333333334 3.333333333333332 3.333333333333334 3.333333333333332h16.666666666666668c1.8399999999999999 0 3.333333333333332-1.4933333333333323 3.333333333333332-3.333333333333332 0-0.9050000000000011-0.3633333333333333-1.7216666666666676-0.9466666666666654-2.3216666666666654l-10.720000000000002-11.011666666666667z' })
                )
            );
        }
    }]);

    return TiMediaEjectOutline;
}(React.Component);

exports.default = TiMediaEjectOutline;
module.exports = exports['default'];