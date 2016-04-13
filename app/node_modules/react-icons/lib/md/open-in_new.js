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

var MdOpenInNew = function (_React$Component) {
    _inherits(MdOpenInNew, _React$Component);

    function MdOpenInNew() {
        _classCallCheck(this, MdOpenInNew);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdOpenInNew).apply(this, arguments));
    }

    _createClass(MdOpenInNew, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm23.36 5h11.64v11.64h-3.3599999999999994v-5.9399999999999995l-16.328333333333333 16.33333333333333-2.3450000000000006-2.346666666666664 16.33-16.328333333333337h-5.938333333333333v-3.3583333333333307z m8.280000000000001 26.64v-11.64h3.3599999999999994v11.64q0 1.3283333333333331-1.0166666666666657 2.3433333333333337t-2.3416666666666686 1.0166666666666657h-23.28333333333333q-1.405000000000002 0-2.3833333333333346-1.0166666666666657t-0.9749999999999996-2.34333333333333v-23.28333333333334q0-1.3266666666666653 0.9766666666666666-2.341666666666665t2.3833333333333346-1.0150000000000006h11.639999999999999v3.3599999999999994h-11.639999999999999v23.283333333333335h23.28333333333334z' })
                )
            );
        }
    }]);

    return MdOpenInNew;
}(React.Component);

exports.default = MdOpenInNew;
module.exports = exports['default'];