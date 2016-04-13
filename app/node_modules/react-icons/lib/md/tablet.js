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

var MdTablet = function (_React$Component) {
    _inherits(MdTablet, _React$Component);

    function MdTablet() {
        _classCallCheck(this, MdTablet);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdTablet).apply(this, arguments));
    }

    _createClass(MdTablet, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm31.640000000000004 30v-20h-23.28333333333334v20h23.283333333333335z m3.359999999999996-23.36q1.3283333333333331 0 2.3433333333333337 1.0166666666666666t1.0166666666666657 2.3416666666666677l-0.0799999999999983 20q0 1.326666666666668-0.9766666666666666 2.3416666666666686t-2.3049999999999997 1.0166666666666657h-29.998333333333335q-1.33 0-2.345-1.0166666666666657t-1.0166666666666664-2.3400000000000034v-20q0-1.33 1.0166666666666668-2.3450000000000006t2.3449999999999998-1.0166666666666657h30z' })
                )
            );
        }
    }]);

    return MdTablet;
}(React.Component);

exports.default = MdTablet;
module.exports = exports['default'];