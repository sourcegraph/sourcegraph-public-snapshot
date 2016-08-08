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

var MdCallEnd = function (_React$Component) {
    _inherits(MdCallEnd, _React$Component);

    function MdCallEnd() {
        _classCallCheck(this, MdCallEnd);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdCallEnd).apply(this, arguments));
    }

    _createClass(MdCallEnd, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm20 15q-4.063333333333333 0-7.656666666666666 1.1716666666666669v5.156666666666666q0 1.1716666666666669-0.9383333333333326 1.5633333333333326-2.5 1.1716666666666669-4.453333333333334 3.046666666666667-0.4666666666666668 0.466666666666665-1.17 0.466666666666665t-1.1716666666666669-0.466666666666665l-4.1433333333333335-4.138333333333332q-0.4666666666666668-0.471666666666664-0.4666666666666668-1.1750000000000007t0.46666666666666673-1.1716666666666669q8.206666666666667-7.811666666666666 19.533333333333335-7.811666666666666t19.53333333333333 7.813333333333334q0.46666666666666856 0.466666666666665 0.46666666666666856 1.1716666666666669t-0.46666666666666856 1.1716666666666669l-4.141666666666666 4.138333333333332q-0.46666666666666856 0.466666666666665-1.1716666666666669 0.466666666666665t-1.1716666666666669-0.466666666666665q-1.9549999999999983-1.875-4.454999999999998-3.046666666666667-0.9383333333333326-0.39000000000000057-0.9383333333333326-1.4833333333333343v-5.156666666666666q-3.905000000000001-1.25-7.655000000000001-1.25z' })
                )
            );
        }
    }]);

    return MdCallEnd;
}(React.Component);

exports.default = MdCallEnd;
module.exports = exports['default'];