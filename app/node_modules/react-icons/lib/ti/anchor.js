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

var TiAnchor = function (_React$Component) {
    _inherits(TiAnchor, _React$Component);

    function TiAnchor() {
        _classCallCheck(this, TiAnchor);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiAnchor).apply(this, arguments));
    }

    _createClass(TiAnchor, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm30 22.5c-0.9216666666666669 0-1.6666666666666679 0.745000000000001-1.6666666666666679 1.6666666666666679 0 4.023333333333333-2.866666666666667 7.390000000000001-6.666666666666668 8.16333333333333v-12.329999999999998h6.666666666666668c0.9216666666666669 0 1.6666666666666679-0.745000000000001 1.6666666666666679-1.6666666666666679s-0.745000000000001-1.6666666666666679-1.6666666666666679-1.6666666666666679h-6.666666666666668v-1.9733333333333292c1.9366666666666674-0.6883333333333326 3.333333333333332-2.5166666666666657 3.333333333333332-4.693333333333333 0-2.7616666666666667-2.2383333333333333-5-5-5s-4.9999999999999964 2.2383333333333315-4.9999999999999964 4.999999999999998c0 2.1750000000000007 1.3966666666666683 4.004999999999999 3.333333333333332 4.693333333333333v1.9733333333333345h-6.666666666666664c-0.9216666666666669 0-1.666666666666666 0.745000000000001-1.666666666666666 1.6666666666666679s0.7449999999999992 1.6666666666666679 1.666666666666666 1.6666666666666679h6.666666666666668v12.330000000000002c-3.8000000000000007-0.7749999999999986-6.666666666666668-4.140000000000001-6.666666666666668-8.163333333333334 0-0.9216666666666669-0.7449999999999992-1.6666666666666679-1.666666666666666-1.6666666666666679s-1.666666666666666 0.745000000000001-1.666666666666666 1.6666666666666679c0 6.433333333333334 5.233333333333334 11.666666666666671 11.666666666666668 11.666666666666671s11.666666666666668-5.233333333333334 11.666666666666668-11.666666666666668c0-0.9216666666666669-0.745000000000001-1.6666666666666679-1.6666666666666679-1.6666666666666679z m-10-14.166666666666666c0.9166666666666679 0 1.6666666666666679 0.75 1.6666666666666679 1.666666666666666s-0.75 1.666666666666666-1.6666666666666679 1.666666666666666-1.6666666666666679-0.75-1.6666666666666679-1.666666666666666 0.75-1.666666666666666 1.6666666666666679-1.666666666666666z' })
                )
            );
        }
    }]);

    return TiAnchor;
}(React.Component);

exports.default = TiAnchor;
module.exports = exports['default'];