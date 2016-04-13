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

var MdLocalPharmacy = function (_React$Component) {
    _inherits(MdLocalPharmacy, _React$Component);

    function MdLocalPharmacy() {
        _classCallCheck(this, MdLocalPharmacy);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdLocalPharmacy).apply(this, arguments));
    }

    _createClass(MdLocalPharmacy, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm26.64 23.36v-3.3599999999999994h-5v-5h-3.2833333333333314v5h-5v3.3599999999999994h5v5h3.2833333333333314v-5h5z m8.36-15v3.283333333333333l-3.3599999999999994 9.999999999999998 3.3599999999999994 10v3.356666666666669h-30v-3.3583333333333343l3.3599999999999994-10-3.3600000000000003-10v-3.283333333333333h21.171666666666667l2.421666666666667-6.716666666666668 3.9066666666666663 1.4833333333333334-1.875 5.236666666666666h4.375z' })
                )
            );
        }
    }]);

    return MdLocalPharmacy;
}(React.Component);

exports.default = MdLocalPharmacy;
module.exports = exports['default'];