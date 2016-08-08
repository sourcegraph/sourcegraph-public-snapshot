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

var MdDoneAll = function (_React$Component) {
    _inherits(MdDoneAll, _React$Component);

    function MdDoneAll() {
        _classCallCheck(this, MdDoneAll);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdDoneAll).apply(this, arguments));
    }

    _createClass(MdDoneAll, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm0.7033333333333334 22.343333333333334l2.3433333333333337-2.3433333333333337 9.296666666666667 9.296666666666667-2.3433333333333337 2.3433333333333373z m36.33-13.046666666666667l2.4200000000000017 2.3433333333333337-20 20-9.375-9.296666666666667 2.421666666666667-2.3433333333333337 6.953333333333333 6.953333333333333z m-7.033333333333331 2.3433333333333337l-10.546666666666667 10.625-2.3433333333333337-2.344999999999999 10.546666666666667-10.625z' })
                )
            );
        }
    }]);

    return MdDoneAll;
}(React.Component);

exports.default = MdDoneAll;
module.exports = exports['default'];