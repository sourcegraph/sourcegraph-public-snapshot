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

var MdReplay = function (_React$Component) {
    _inherits(MdReplay, _React$Component);

    function MdReplay() {
        _classCallCheck(this, MdReplay);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdReplay).apply(this, arguments));
    }

    _createClass(MdReplay, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm20 8.360000000000001q5.546666666666667 0 9.453333333333333 3.9066666666666663t3.9066666666666663 9.374999999999998q0 5.546666666666667-3.9450000000000003 9.453333333333333t-9.415 3.905000000000001-9.411666666666667-3.905000000000001-3.9450000000000003-9.453333333333333h3.360000000000001q0 4.140000000000001 2.9299999999999997 7.07t7.07 2.9299999999999997 7.07-2.9299999999999997 2.9299999999999997-7.07-2.9299999999999997-7.07-7.07-2.9299999999999997v6.716666666666669l-8.360000000000001-8.358333333333334 8.360000000000001-8.36v6.716666666666667z' })
                )
            );
        }
    }]);

    return MdReplay;
}(React.Component);

exports.default = MdReplay;
module.exports = exports['default'];