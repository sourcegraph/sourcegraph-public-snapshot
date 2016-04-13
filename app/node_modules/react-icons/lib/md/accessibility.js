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

var MdAccessibility = function (_React$Component) {
    _inherits(MdAccessibility, _React$Component);

    function MdAccessibility() {
        _classCallCheck(this, MdAccessibility);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdAccessibility).apply(this, arguments));
    }

    _createClass(MdAccessibility, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm35 15h-10v21.64h-3.3599999999999994v-10h-3.2833333333333314v10h-3.356666666666669v-21.64h-10v-3.3599999999999994h30v3.3599999999999994z m-15-11.64q1.3283333333333331 0 2.3433333333333337 0.9766666666666666t1.0166666666666657 2.3049999999999997-1.0166666666666657 2.3433333333333337-2.3433333333333337 1.0150000000000006-2.3433333333333337-1.0166666666666657-1.0166666666666657-2.341666666666667 1.0166666666666657-2.3049999999999997 2.3433333333333337-0.9766666666666675z' })
                )
            );
        }
    }]);

    return MdAccessibility;
}(React.Component);

exports.default = MdAccessibility;
module.exports = exports['default'];