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

var MdVideocam = function (_React$Component) {
    _inherits(MdVideocam, _React$Component);

    function MdVideocam() {
        _classCallCheck(this, MdVideocam);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdVideocam).apply(this, arguments));
    }

    _createClass(MdVideocam, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm28.36 17.5l6.640000000000001-6.639999999999999v18.28333333333334l-6.640000000000001-6.643333333333338v5.861666666666665q0 0.7033333333333331-0.5083333333333329 1.1716666666666669t-1.2100000000000009 0.466666666666665h-20q-0.7033333333333331 0-1.1716666666666669-0.466666666666665t-0.46999999999999886-1.1716666666666633v-16.71666666666667q0-0.7049999999999983 0.4666666666666668-1.173333333333332t1.1733333333333338-0.4666666666666668h20q0.7033333333333331 0 1.211666666666666 0.4666666666666668t0.5100000000000016 1.1716666666666669v5.859999999999999z' })
                )
            );
        }
    }]);

    return MdVideocam;
}(React.Component);

exports.default = MdVideocam;
module.exports = exports['default'];