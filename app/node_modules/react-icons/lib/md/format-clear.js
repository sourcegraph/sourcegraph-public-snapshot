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

var MdFormatClear = function (_React$Component) {
    _inherits(MdFormatClear, _React$Component);

    function MdFormatClear() {
        _classCallCheck(this, MdFormatClear);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdFormatClear).apply(this, arguments));
    }

    _createClass(MdFormatClear, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm10 8.360000000000001h23.36v5h-9.688333333333333l-2.6566666666666663 6.249999999999998-3.5166666666666657-3.4383333333333326 1.1716666666666669-2.8133333333333326h-3.9833333333333343l-4.6899999999999995-4.688333333333334v-0.3133333333333326z m-4.533333333333334 0l0.46999999999999975 0.39000000000000057 24.063333333333333 24.140000000000008-2.109999999999996 2.1099999999999923-9.453333333333333-9.453333333333333-2.578333333333333 6.093333333333334h-5l4.063333333333333-9.61-11.563333333333336-11.563333333333334z' })
                )
            );
        }
    }]);

    return MdFormatClear;
}(React.Component);

exports.default = MdFormatClear;
module.exports = exports['default'];