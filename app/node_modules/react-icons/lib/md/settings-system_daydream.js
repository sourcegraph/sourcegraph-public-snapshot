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

var MdSettingsSystemDaydream = function (_React$Component) {
    _inherits(MdSettingsSystemDaydream, _React$Component);

    function MdSettingsSystemDaydream() {
        _classCallCheck(this, MdSettingsSystemDaydream);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdSettingsSystemDaydream).apply(this, arguments));
    }

    _createClass(MdSettingsSystemDaydream, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm35 31.71666666666667v-23.433333333333337h-30v23.433333333333337h30z m0-26.71666666666667q1.3283333333333331 0 2.3433333333333337 1.0166666666666666t1.0166666666666657 2.3416666666666677v23.283333333333335q0 1.3266666666666644-1.0166666666666657 2.341666666666665t-2.3433333333333337 1.0166666666666657h-30q-1.3283333333333331 0-2.3433333333333333-1.0166666666666657t-1.0166666666666666-2.3433333333333337v-23.28333333333333q0-1.3266666666666689 1.0166666666666666-2.3416666666666686t2.3433333333333333-1.0150000000000006h30z m-20 21.64q-2.033333333333333 0-3.5166666666666657-1.4833333333333343t-1.4833333333333343-3.518333333333331q0-1.9533333333333331 1.288333333333334-3.3599999999999994t3.163333333333334-1.5633333333333326h0.31666666666666643q1.6350000000000016-3.358333333333338 5.231666666666666-3.358333333333338 2.1883333333333326 0 3.828333333333333 1.4450000000000003t1.9533333333333331 3.5566666666666684h0.07666666666666799q1.7166666666666686 0 2.9299999999999997 1.211666666666666t1.211666666666666 2.9299999999999997-1.2100000000000009 2.9283333333333346-2.9299999999999997 1.211666666666666h-10.86z' })
                )
            );
        }
    }]);

    return MdSettingsSystemDaydream;
}(React.Component);

exports.default = MdSettingsSystemDaydream;
module.exports = exports['default'];