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

var MdCheckBox = function (_React$Component) {
    _inherits(MdCheckBox, _React$Component);

    function MdCheckBox() {
        _classCallCheck(this, MdCheckBox);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdCheckBox).apply(this, arguments));
    }

    _createClass(MdCheckBox, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm16.64 28.36l15-15-2.3433333333333337-2.421666666666667-12.656666666666666 12.656666666666666-5.938333333333333-5.938333333333333-2.341666666666667 2.3433333333333337z m15-23.36q1.4066666666666663 0 2.383333333333333 1.0166666666666666t0.9766666666666666 2.3400000000000007v23.28333333333333q0 1.326666666666668-0.9766666666666666 2.3416666666666686t-2.383333333333333 1.0166666666666657h-23.28333333333333q-1.405000000000002 0-2.3833333333333346-1.0166666666666657t-0.9733333333333345-2.341666666666665v-23.28333333333334q0-1.3266666666666653 0.9749999999999996-2.341666666666665t2.383333333333333-1.0150000000000006h23.28333333333333z' })
                )
            );
        }
    }]);

    return MdCheckBox;
}(React.Component);

exports.default = MdCheckBox;
module.exports = exports['default'];