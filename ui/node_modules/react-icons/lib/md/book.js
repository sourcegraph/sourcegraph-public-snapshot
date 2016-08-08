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

var MdBook = function (_React$Component) {
    _inherits(MdBook, _React$Component);

    function MdBook() {
        _classCallCheck(this, MdBook);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdBook).apply(this, arguments));
    }

    _createClass(MdBook, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm10 6.640000000000001v13.36l4.140000000000001-2.5 4.216666666666669 2.5v-13.36h-8.35666666666667z m20-3.2800000000000002q1.3283333333333331 0 2.3433333333333337 0.9766666666666666t1.0166666666666657 2.3050000000000006v26.71666666666667q0 1.3299999999999983-1.0166666666666657 2.306666666666665t-2.3433333333333337 0.9750000000000085h-20q-1.3283333333333331 0-2.3433333333333337-0.9766666666666666t-1.0166666666666666-2.306666666666665v-26.713333333333342q0-1.330000000000001 1.0166666666666666-2.3066666666666675t2.3433333333333337-0.9766666666666666h20z' })
                )
            );
        }
    }]);

    return MdBook;
}(React.Component);

exports.default = MdBook;
module.exports = exports['default'];