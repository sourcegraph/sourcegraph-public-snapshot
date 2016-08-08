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

var MdPanorama = function (_React$Component) {
    _inherits(MdPanorama, _React$Component);

    function MdPanorama() {
        _classCallCheck(this, MdPanorama);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdPanorama).apply(this, arguments));
    }

    _createClass(MdPanorama, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm14.14 20.86l-5.783333333333335 7.5h23.283333333333335l-7.5-10-5.783333333333335 7.5z m24.22 9.14q0 1.3283333333333331-1.0166666666666657 2.3433333333333337t-2.3416666666666686 1.0166666666666657h-30.001666666666665q-1.326666666666667 0-2.341666666666667-1.0166666666666657t-1.0166666666666666-2.3433333333333337v-20q0-1.3283333333333331 1.0166666666666666-2.3433333333333337t2.341666666666667-1.0166666666666657h30q1.3299999999999983 0 2.344999999999999 1.0166666666666666t1.0166666666666657 2.343333333333333v20z' })
                )
            );
        }
    }]);

    return MdPanorama;
}(React.Component);

exports.default = MdPanorama;
module.exports = exports['default'];