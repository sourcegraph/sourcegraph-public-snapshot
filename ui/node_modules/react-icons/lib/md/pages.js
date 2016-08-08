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

var MdPages = function (_React$Component) {
    _inherits(MdPages, _React$Component);

    function MdPages() {
        _classCallCheck(this, MdPages);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdPages).apply(this, arguments));
    }

    _createClass(MdPages, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm31.640000000000004 5q1.3283333333333367 0 2.34333333333333 1.0166666666666666t1.0166666666666657 2.3400000000000007v10.000000000000002h-8.361666666666668l1.716666666666665-6.716666666666669-6.716666666666669 1.7166666666666668v-8.356666666666667h10z m-3.280000000000001 23.36l-1.716666666666665-6.716666666666669h8.356666666666662v10q0 1.326666666666668-1.0133333333333354 2.3416666666666686t-2.3433333333333337 1.0166666666666657h-10v-8.36z m-15-6.719999999999999l-1.7166666666666668 6.716666666666669 6.716666666666667-1.7166666666666686v8.36h-10q-1.3283333333333331 0-2.3433333333333337-1.0166666666666657t-1.0166666666666693-2.34v-10h8.361666666666668z m-8.360000000000003-13.28q0-1.3283333333333331 1.0166666666666666-2.3433333333333337t2.3400000000000007-1.0166666666666675h10.000000000000002v8.361666666666668l-6.716666666666669-1.7166666666666668 1.7166666666666668 6.716666666666667h-8.356666666666667v-10z' })
                )
            );
        }
    }]);

    return MdPages;
}(React.Component);

exports.default = MdPages;
module.exports = exports['default'];