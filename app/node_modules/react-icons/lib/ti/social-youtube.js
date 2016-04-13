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

var TiSocialYoutube = function (_React$Component) {
    _inherits(TiSocialYoutube, _React$Component);

    function TiSocialYoutube() {
        _classCallCheck(this, TiSocialYoutube);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiSocialYoutube).apply(this, arguments));
    }

    _createClass(TiSocialYoutube, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm38 14.3c-0.29999999999999716-2.5-0.7000000000000028-4.300000000000001-1.7000000000000028-5-1-0.8000000000000007-9.600000000000001-1-16.3-1s-15.3 0.1999999999999993-16.3 1c-1 0.6999999999999993-1.4 2.5-1.7 5s-0.30000000000000004 4-0.30000000000000004 5.699999999999999 0 3.1999999999999993 0.30000000000000004 5.699999999999999 0.7 4.300000000000001 1.7 5c1 0.8000000000000007 9.6 1 16.3 1 6.699999999999999 0 15.3-0.1999999999999993 16.3-1 1-0.6999999999999993 1.3999999999999986-2.5 1.7000000000000028-5s0.29999999999999716-4 0.29999999999999716-5.699999999999999 0-3.1999999999999993-0.29999999999999716-5.699999999999999z m-21.3 11.7v-12l10 6-10 6z' })
                )
            );
        }
    }]);

    return TiSocialYoutube;
}(React.Component);

exports.default = TiSocialYoutube;
module.exports = exports['default'];