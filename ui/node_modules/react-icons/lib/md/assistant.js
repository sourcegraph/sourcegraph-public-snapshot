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

var MdAssistant = function (_React$Component) {
    _inherits(MdAssistant, _React$Component);

    function MdAssistant() {
        _classCallCheck(this, MdAssistant);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdAssistant).apply(this, arguments));
    }

    _createClass(MdAssistant, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm23.125 21.483333333333334l6.875-3.121666666666666-6.875-3.128333333333334-3.125-6.871666666666666-3.125 6.871666666666666-6.875 3.128333333333334 6.875 3.125 3.125 6.875z m8.516666666666666-18.123333333333335q1.326666666666668 0 2.3416666666666686 0.9766666666666666t1.0166666666666657 2.3050000000000006v23.358333333333334q0 1.3299999999999983-1.0166666666666657 2.344999999999999t-2.3433333333333337 1.0166666666666657h-6.640000000000001l-5 5-5-5h-6.639999999999999q-1.3283333333333331 0-2.3433333333333337-1.0166666666666657t-1.0166666666666675-2.344999999999999v-23.356666666666666q0-1.328333333333334 1.0166666666666666-2.3050000000000015t2.343333333333333-0.9766666666666666h23.283333333333335z' })
                )
            );
        }
    }]);

    return MdAssistant;
}(React.Component);

exports.default = MdAssistant;
module.exports = exports['default'];