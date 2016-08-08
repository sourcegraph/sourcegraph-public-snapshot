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

var MdSnooze = function (_React$Component) {
    _inherits(MdSnooze, _React$Component);

    function MdSnooze() {
        _classCallCheck(this, MdSnooze);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdSnooze).apply(this, arguments));
    }

    _createClass(MdSnooze, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm15 18.36v-3.3599999999999994h10v2.9666666666666686l-6.016666666666666 7.033333333333331h6.016666666666666v3.3599999999999994h-10v-3.046666666666667l6.016666666666666-6.953333333333333h-6.016666666666666z m5 15q4.843333333333334 0 8.241666666666667-3.4383333333333326t3.3999999999999986-8.283333333333331-3.3999999999999986-8.24-8.241666666666667-3.398333333333335-8.241666666666667 3.3999999999999986-3.4000000000000004 8.240000000000002 3.4000000000000004 8.283333333333331 8.241666666666667 3.4350000000000023z m0-26.720000000000002q6.25 0 10.625 4.413333333333334t4.375 10.58666666666667-4.375 10.586666666666666-10.625 4.413333333333341-10.625-4.413333333333334-4.375-10.586666666666673 4.375-10.586666666666668 10.625-4.413333333333332z m16.64 2.8916666666666675l-2.1099999999999994 2.576666666666666-7.656666666666666-6.483333333333333 2.1099999999999994-2.5z m-23.516666666666666-3.9083333333333306l-7.653333333333334 6.409999999999999-2.1083333333333334-2.5 7.655-6.41z' })
                )
            );
        }
    }]);

    return MdSnooze;
}(React.Component);

exports.default = MdSnooze;
module.exports = exports['default'];