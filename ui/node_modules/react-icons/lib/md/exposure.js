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

var MdExposure = function (_React$Component) {
    _inherits(MdExposure, _React$Component);

    function MdExposure() {
        _classCallCheck(this, MdExposure);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdExposure).apply(this, arguments));
    }

    _createClass(MdExposure, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm33.36 33.36v-26.716666666666665l-26.71666666666667 26.716666666666665h26.71666666666667z m-25-25v3.283333333333333h10v-3.283333333333333h-10z m25-5q1.3283333333333331 0 2.3049999999999997 0.9766666666666666t0.9750000000000014 2.3049999999999997v26.71666666666667q0 1.3299999999999983-0.9766666666666666 2.306666666666665t-2.306666666666665 0.9750000000000014h-26.713333333333335q-1.330000000000001 0-2.3066666666666675-0.9766666666666666t-0.9766666666666666-2.306666666666665v-26.713333333333335q0-1.330000000000001 0.9766666666666666-2.3066666666666675t2.3050000000000006-0.9766666666666666h26.71666666666667z m-8.36 25h-3.3599999999999994v-3.3599999999999994h3.3599999999999994v-3.3599999999999994h3.3599999999999994v3.3599999999999994h3.2833333333333314v3.3599999999999994h-3.2833333333333314v3.2833333333333314h-3.3599999999999994v-3.2833333333333314z' })
                )
            );
        }
    }]);

    return MdExposure;
}(React.Component);

exports.default = MdExposure;
module.exports = exports['default'];