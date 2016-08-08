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

var MdHistory = function (_React$Component) {
    _inherits(MdHistory, _React$Component);

    function MdHistory() {
        _classCallCheck(this, MdHistory);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdHistory).apply(this, arguments));
    }

    _createClass(MdHistory, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm20 13.360000000000001h2.5v7.033333333333333l5.859999999999999 3.513333333333332-1.25 2.0333333333333314-7.109999999999999-4.301666666666662v-8.276666666666669z m1.6400000000000006-8.360000000000001q6.171666666666667 0 10.586666666666666 4.375t4.413333333333341 10.625-4.413333333333334 10.625-10.586666666666673 4.375-10.546666666666667-4.375l2.3433333333333337-2.421666666666667q3.4383333333333326 3.4383333333333326 8.203333333333333 3.4383333333333326 4.843333333333334 0 8.283333333333331-3.3999999999999986t3.4366666666666674-8.240000000000002-3.4383333333333326-8.24-8.283333333333331-3.4000000000000004-8.24 3.4000000000000004-3.398333333333335 8.238333333333335h5l-6.716666666666669 6.719999999999999-0.15833333333333321-0.23333333333333428-6.483333333333332-6.486666666666665h5q0-6.25 4.411666666666666-10.625t10.586666666666668-4.375z' })
                )
            );
        }
    }]);

    return MdHistory;
}(React.Component);

exports.default = MdHistory;
module.exports = exports['default'];