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

var TiMediaRewind = function (_React$Component) {
    _inherits(TiMediaRewind, _React$Component);

    function TiMediaRewind() {
        _classCallCheck(this, TiMediaRewind);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiMediaRewind).apply(this, arguments));
    }

    _createClass(TiMediaRewind, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm17 10.666666666666668c-0.8133333333333326 0-1.5500000000000007 0.32833333333333314-2.088333333333333 0.8533333333333335-3.966666666666667 3.8583333333333325-9.911666666666667 9.65-9.911666666666667 9.65l9.906666666666666 9.646666666666668c0.543333333333333 0.5249999999999986 1.2800000000000011 0.8500000000000014 2.0933333333333337 0.8500000000000014 1.6566666666666663 0 3-1.3416666666666686 3-3v-15c0-1.6549999999999994-1.3433333333333337-3-3-3z m15 0c-0.8133333333333326 0-1.5500000000000007 0.32833333333333314-2.0883333333333347 0.8533333333333335-3.966666666666665 3.8583333333333325-9.911666666666665 9.65-9.911666666666665 9.65l9.906666666666666 9.646666666666668c0.543333333333333 0.5249999999999986 1.2800000000000011 0.8500000000000014 2.0933333333333337 0.8500000000000014 1.6566666666666663 0 3-1.3416666666666686 3-3v-15c0-1.6549999999999994-1.3433333333333337-3-3-3z' })
                )
            );
        }
    }]);

    return TiMediaRewind;
}(React.Component);

exports.default = TiMediaRewind;
module.exports = exports['default'];