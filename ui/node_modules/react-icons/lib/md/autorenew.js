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

var MdAutorenew = function (_React$Component) {
    _inherits(MdAutorenew, _React$Component);

    function MdAutorenew() {
        _classCallCheck(this, MdAutorenew);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdAutorenew).apply(this, arguments));
    }

    _createClass(MdAutorenew, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm31.25 12.89q2.1099999999999994 3.203333333333333 2.1099999999999994 7.109999999999999 0 5.466666666666669-3.9450000000000003 9.413333333333334t-9.415 3.9450000000000003v5l-6.638333333333334-6.716666666666669 6.638333333333334-6.641666666666666v5q4.141666666666666 0 7.071666666666665-2.9299999999999997t2.9283333333333346-7.07q0-2.3433333333333337-1.1700000000000017-4.688333333333333z m-11.25-2.8900000000000006q-4.139999999999999 0-7.07 2.9299999999999997t-2.9299999999999997 7.07q0 2.578333333333333 1.1716666666666669 4.688333333333333l-2.421666666666667 2.421666666666667q-2.1099999999999994-3.203333333333333-2.1099999999999994-7.109999999999999 0-5.466666666666667 3.9450000000000003-9.413333333333334t9.415-3.9449999999999994v-5l6.638333333333335 6.716666666666668-6.638333333333335 6.641666666666666v-5z' })
                )
            );
        }
    }]);

    return MdAutorenew;
}(React.Component);

exports.default = MdAutorenew;
module.exports = exports['default'];