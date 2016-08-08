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

var TiSocialFacebook = function (_React$Component) {
    _inherits(TiSocialFacebook, _React$Component);

    function TiSocialFacebook() {
        _classCallCheck(this, TiSocialFacebook);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiSocialFacebook).apply(this, arguments));
    }

    _createClass(TiSocialFacebook, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm21.666666666666668 16.666666666666668h5v5h-5v11.666666666666668h-5v-11.666666666666668h-5v-5h5v-2.0916666666666686c0-1.9833333333333343 0.6233333333333348-4.483333333333334 1.8633333333333333-5.8533333333333335 1.2399999999999984-1.3716666666666653 2.7866666666666653-2.054999999999999 4.643333333333334-2.054999999999999h3.4933333333333323v5.000000000000001h-3.5c-0.8300000000000018 0-1.5 0.6699999999999999-1.5 1.5v3.5z' })
                )
            );
        }
    }]);

    return TiSocialFacebook;
}(React.Component);

exports.default = TiSocialFacebook;
module.exports = exports['default'];