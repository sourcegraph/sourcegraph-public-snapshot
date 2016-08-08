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

var TiThList = function (_React$Component) {
    _inherits(TiThList, _React$Component);

    function TiThList() {
        _classCallCheck(this, TiThList);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiThList).apply(this, arguments));
    }

    _createClass(TiThList, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm31.666666666666668 28.333333333333336h-11.666666666666668c-1.8383333333333347 0-3.333333333333332 1.495000000000001-3.333333333333332 3.333333333333332s1.495000000000001 3.333333333333332 3.333333333333332 3.333333333333332h11.666666666666668c1.8383333333333347 0 3.333333333333332-1.4949999999999974 3.333333333333332-3.333333333333332s-1.4949999999999974-3.333333333333332-3.333333333333332-3.333333333333332z m0-11.666666666666668h-11.666666666666668c-1.8383333333333347 0-3.333333333333332 1.495000000000001-3.333333333333332 3.333333333333332s1.495000000000001 3.333333333333332 3.333333333333332 3.333333333333332h11.666666666666668c1.8383333333333347 0 3.333333333333332-1.495000000000001 3.333333333333332-3.333333333333332s-1.4949999999999974-3.333333333333332-3.333333333333332-3.333333333333332z m0-11.666666666666668h-11.666666666666668c-1.8383333333333347 0-3.333333333333332 1.495-3.333333333333332 3.333333333333334s1.495000000000001 3.333333333333334 3.333333333333332 3.333333333333334h11.666666666666668c1.8383333333333347 0 3.333333333333332-1.495000000000001 3.333333333333332-3.333333333333334s-1.4949999999999974-3.333333333333334-3.333333333333332-3.333333333333334z m-19.166666666666668 26.666666666666668c0 2.3000000000000007-1.8666666666666671 4.166666666666668-4.166666666666668 4.166666666666668s-4.166666666666665-1.8666666666666671-4.166666666666665-4.166666666666668 1.8666666666666671-4.166666666666668 4.166666666666667-4.166666666666668 4.166666666666666 1.8666666666666671 4.166666666666666 4.166666666666668z m0-11.666666666666668c0 2.3000000000000007-1.8666666666666671 4.166666666666668-4.166666666666668 4.166666666666668s-4.166666666666665-1.8666666666666671-4.166666666666665-4.166666666666668 1.8666666666666671-4.166666666666666 4.166666666666667-4.166666666666666 4.166666666666666 1.8666666666666654 4.166666666666666 4.166666666666666z m0-11.666666666666666c0 2.299999999999999-1.8666666666666671 4.166666666666666-4.166666666666668 4.166666666666666s-4.166666666666665-1.8666666666666671-4.166666666666665-4.166666666666666 1.8666666666666671-4.166666666666667 4.166666666666667-4.166666666666667 4.166666666666666 1.8666666666666671 4.166666666666666 4.166666666666667z' })
                )
            );
        }
    }]);

    return TiThList;
}(React.Component);

exports.default = TiThList;
module.exports = exports['default'];