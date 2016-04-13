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

var TiCamera = function (_React$Component) {
    _inherits(TiCamera, _React$Component);

    function TiCamera() {
        _classCallCheck(this, TiCamera);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiCamera).apply(this, arguments));
    }

    _createClass(TiCamera, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm31.666666666666668 10h-2.6433333333333344l-1.6666666666666679-1.666666666666666c-0.966666666666665-0.9666666666666668-2.658333333333335-1.666666666666667-4.023333333333333-1.666666666666667h-6.666666666666668c-1.3666666666666671 0-3.0583333333333336 0.7000000000000002-4.023333333333333 1.666666666666667l-1.666666666666666 1.666666666666666h-2.643333333333331c-2.756666666666667 0-5 2.243333333333334-5 5v13.333333333333336c-4.440892098500626e-16 2.7566666666666677 2.243333333333333 5 5 5h23.333333333333336c2.7566666666666677 0 5-2.2433333333333323 5-5v-13.333333333333336c0-2.7566666666666677-2.2433333333333323-5-5-5z m-11.666666666666668 16.666666666666668c-3.2216666666666676 0-5.833333333333334-2.6133333333333333-5.833333333333334-5.833333333333332 1.7763568394002505e-15-3.2233333333333327 2.61166666666667-5.833333333333336 5.833333333333334-5.833333333333336s5.833333333333336 2.6099999999999994 5.833333333333336 5.833333333333336c0 3.219999999999999-2.611666666666668 5.833333333333336-5.833333333333336 5.833333333333336z m10-7.833333333333332c-1.1999999999999993 0-2.166666666666668-0.966666666666665-2.166666666666668-2.166666666666668s0.966666666666665-2.166666666666668 2.166666666666668-2.166666666666668 2.1666666666666643 0.9666666666666668 2.1666666666666643 2.166666666666668-0.966666666666665 2.166666666666668-2.166666666666668 2.166666666666668z' })
                )
            );
        }
    }]);

    return TiCamera;
}(React.Component);

exports.default = TiCamera;
module.exports = exports['default'];