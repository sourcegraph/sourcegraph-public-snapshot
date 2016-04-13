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

var TiThLargeOutline = function (_React$Component) {
    _inherits(TiThLargeOutline, _React$Component);

    function TiThLargeOutline() {
        _classCallCheck(this, TiThLargeOutline);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiThLargeOutline).apply(this, arguments));
    }

    _createClass(TiThLargeOutline, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm15 3.3333333333333335h-8.333333333333332c-1.8383333333333347 0-3.3333333333333344 1.4933333333333336-3.3333333333333344 3.3333333333333335v8.333333333333332c0 1.8399999999999999 1.4949999999999997 3.333333333333332 3.3333333333333335 3.333333333333332h8.333333333333332c1.8383333333333347 0 3.333333333333332-1.4933333333333323 3.333333333333332-3.333333333333334v-8.333333333333332c0-1.8399999999999999-1.495000000000001-3.3333333333333326-3.333333333333334-3.3333333333333326z m0 11.666666666666668h-8.333333333333332v-8.333333333333336h8.333333333333332v8.333333333333334z m18.333333333333336-11.666666666666668h-8.333333333333336c-1.8399999999999999-4.440892098500626e-16-3.333333333333332 1.4933333333333332-3.333333333333332 3.333333333333333v8.333333333333332c0 1.8399999999999999 1.4933333333333323 3.333333333333332 3.333333333333332 3.333333333333332h8.333333333333336c1.8400000000000034 0 3.3333333333333357-1.4933333333333323 3.3333333333333357-3.333333333333334v-8.333333333333332c0-1.8399999999999999-1.4933333333333323-3.3333333333333326-3.3333333333333357-3.3333333333333326z m0 11.666666666666668h-8.333333333333336v-8.333333333333336h8.333333333333336v8.333333333333334z m-18.333333333333336 6.666666666666666h-8.333333333333332c-1.8383333333333347 0-3.3333333333333344 1.4933333333333323-3.3333333333333344 3.333333333333332v8.333333333333336c0 1.8400000000000034 1.4949999999999997 3.3333333333333357 3.3333333333333335 3.3333333333333357h8.333333333333332c1.8383333333333347 0 3.333333333333332-1.4933333333333323 3.333333333333332-3.3333333333333357v-8.333333333333336c0-1.8399999999999999-1.495000000000001-3.333333333333332-3.333333333333334-3.333333333333332z m0 11.666666666666668h-8.333333333333332v-8.333333333333336h8.333333333333332v8.333333333333336z m18.333333333333336-11.666666666666668h-8.333333333333336c-1.8399999999999999 0-3.333333333333332 1.4933333333333323-3.333333333333332 3.333333333333332v8.333333333333336c0 1.8400000000000034 1.4933333333333323 3.3333333333333357 3.333333333333332 3.3333333333333357h8.333333333333336c1.8400000000000034 0 3.3333333333333357-1.4933333333333323 3.3333333333333357-3.3333333333333357v-8.333333333333336c0-1.8399999999999999-1.4933333333333323-3.333333333333332-3.3333333333333357-3.333333333333332z m0 11.666666666666668h-8.333333333333336v-8.333333333333336h8.333333333333336v8.333333333333336z' })
                )
            );
        }
    }]);

    return TiThLargeOutline;
}(React.Component);

exports.default = TiThLargeOutline;
module.exports = exports['default'];