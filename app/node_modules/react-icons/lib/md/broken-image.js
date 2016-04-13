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

var MdBrokenImage = function (_React$Component) {
    _inherits(MdBrokenImage, _React$Component);

    function MdBrokenImage() {
        _classCallCheck(this, MdBrokenImage);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdBrokenImage).apply(this, arguments));
    }

    _createClass(MdBrokenImage, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm30 19.063333333333336l5 5v7.578333333333333q0 1.3283333333333367-1.0166666666666657 2.34333333333333t-2.3416666666666686 1.0166666666666657h-23.28333333333333q-1.3266666666666689 0-2.3416666666666686-1.0166666666666657t-1.0166666666666657-2.341666666666665v-10.938333333333333l5 5 6.641666666666666-6.716666666666669 6.716666666666669 6.716666666666669z m5-10.703333333333333v10.938333333333333l-5-5-6.640000000000001 6.716666666666669-6.716666666666669-6.716666666666669-6.643333333333331 6.71833333333333-5-5.078333333333331v-7.578333333333333q0-1.3283333333333331 1.0166666666666666-2.3433333333333337t2.3433333333333346-1.0166666666666675h23.28333333333334q1.326666666666668 0 2.3416666666666686 1.0166666666666666t1.0149999999999935 2.3433333333333346z' })
                )
            );
        }
    }]);

    return MdBrokenImage;
}(React.Component);

exports.default = MdBrokenImage;
module.exports = exports['default'];