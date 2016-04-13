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

var MdCrop = function (_React$Component) {
    _inherits(MdCrop, _React$Component);

    function MdCrop() {
        _classCallCheck(this, MdCrop);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdCrop).apply(this, arguments));
    }

    _createClass(MdCrop, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm11.64 28.36h26.71666666666667v3.2833333333333314h-6.716666666666669v6.716666666666669h-3.2833333333333314v-6.716666666666669h-16.715000000000003q-1.3283333333333314 0-2.304999999999998-0.9783333333333317t-0.9749999999999996-2.306666666666665v-16.715000000000003h-6.7200000000000015v-3.2833333333333314h6.716666666666668v-6.716666666666666h3.283333333333333v26.71666666666667z m16.72-3.3599999999999994v-13.36h-13.36v-3.283333333333333h13.36q1.3283333333333331 0 2.3049999999999997 0.9783333333333335t0.9750000000000014 2.3066666666666666v13.358333333333333h-3.2833333333333314z' })
                )
            );
        }
    }]);

    return MdCrop;
}(React.Component);

exports.default = MdCrop;
module.exports = exports['default'];