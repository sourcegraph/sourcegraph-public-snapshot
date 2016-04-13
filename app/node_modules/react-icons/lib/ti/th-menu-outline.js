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

var TiThMenuOutline = function (_React$Component) {
    _inherits(TiThMenuOutline, _React$Component);

    function TiThMenuOutline() {
        _classCallCheck(this, TiThMenuOutline);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiThMenuOutline).apply(this, arguments));
    }

    _createClass(TiThMenuOutline, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm31.666666666666668 30c0.9166666666666679 0 1.6666666666666679 0.75 1.6666666666666679 1.6666666666666679s-0.75 1.6666666666666679-1.6666666666666679 1.6666666666666679h-23.333333333333336c-0.9166666666666652 0-1.6666666666666652-0.75-1.6666666666666652-1.6666666666666679s0.75-1.6666666666666679 1.666666666666667-1.6666666666666679h23.333333333333336z m0-3.333333333333332h-23.333333333333336c-2.756666666666665 0-4.999999999999998 2.2433333333333323-4.999999999999998 5s2.243333333333334 5.0000000000000036 5 5.0000000000000036h23.333333333333336c2.7566666666666677 0 5-2.2433333333333323 5-5s-2.2433333333333323-5-5-5z m0-8.333333333333332c0.9166666666666679 0 1.6666666666666679 0.75 1.6666666666666679 1.6666666666666679s-0.75 1.6666666666666679-1.6666666666666679 1.6666666666666679h-23.333333333333336c-0.9166666666666652 0-1.6666666666666652-0.75-1.6666666666666652-1.6666666666666679s0.75-1.6666666666666679 1.666666666666667-1.6666666666666679h23.333333333333336z m0-3.333333333333334h-23.333333333333336c-2.756666666666665 0-4.999999999999998 2.243333333333334-4.999999999999998 4.999999999999998s2.243333333333334 5 5 5h23.333333333333336c2.7566666666666677 0 5-2.2433333333333323 5-5s-2.2433333333333323-5-5-5z m0-8.333333333333336c0.9166666666666679 8.881784197001252e-16 1.6666666666666679 0.7500000000000009 1.6666666666666679 1.6666666666666679s-0.75 1.666666666666666-1.6666666666666679 1.666666666666666h-23.333333333333336c-0.9166666666666652 0-1.6666666666666652-0.75-1.6666666666666652-1.666666666666666s0.75-1.666666666666667 1.666666666666667-1.666666666666667h23.333333333333336z m0-3.3333333333333326h-23.333333333333336c-2.756666666666665 0-4.999999999999998 2.2433333333333336-4.999999999999998 5s2.243333333333333 5 5 5h23.333333333333336c2.7566666666666677 0 5-2.243333333333334 5-5s-2.2433333333333323-5-5-5z' })
                )
            );
        }
    }]);

    return TiThMenuOutline;
}(React.Component);

exports.default = TiThMenuOutline;
module.exports = exports['default'];