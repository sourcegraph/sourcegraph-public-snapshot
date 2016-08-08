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

var MdLocalPizza = function (_React$Component) {
    _inherits(MdLocalPizza, _React$Component);

    function MdLocalPizza() {
        _classCallCheck(this, MdLocalPizza);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdLocalPizza).apply(this, arguments));
    }

    _createClass(MdLocalPizza, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm20 25q1.3283333333333331 0 2.3433333333333337-1.0166666666666657t1.0166666666666657-2.3416666666666686-1.0166666666666657-2.3049999999999997-2.3433333333333337-0.9766666666666666-2.3433333333333337 0.9766666666666666-1.0166666666666657 2.3049999999999997 1.0166666666666657 2.3433333333333337 2.3433333333333337 1.0150000000000006z m-8.36-13.36q0 1.3283333333333331 1.0166666666666657 2.3433333333333337t2.34 1.0166666666666657 2.341666666666665-1.0166666666666657 1.0166666666666657-2.3433333333333337-1.0166666666666657-2.3049999999999997-2.338333333333331-0.9749999999999996-2.3466666666666676 0.9733333333333327-1.0166666666666657 2.3066666666666666z m8.36-8.280000000000001q8.983333333333334 8.881784197001252e-16 15 6.640000000000001l-15 26.64-15-26.64q6.016666666666666-6.640000000000001 15-6.640000000000001z' })
                )
            );
        }
    }]);

    return MdLocalPizza;
}(React.Component);

exports.default = MdLocalPizza;
module.exports = exports['default'];