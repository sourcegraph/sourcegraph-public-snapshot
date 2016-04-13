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

var MdUnfoldMore = function (_React$Component) {
    _inherits(MdUnfoldMore, _React$Component);

    function MdUnfoldMore() {
        _classCallCheck(this, MdUnfoldMore);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdUnfoldMore).apply(this, arguments));
    }

    _createClass(MdUnfoldMore, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm20 30.313333333333333l5.313333333333336-5.313333333333333 2.3433333333333337 2.3433333333333337-7.65666666666667 7.656666666666666-7.656666666666666-7.656666666666666 2.343333333333332-2.3433333333333337z m0-20.625l-5.313333333333334 5.311666666666667-2.343333333333332-2.34 7.656666666666666-7.66 7.656666666666666 7.658333333333333-2.34333333333333 2.341666666666667z' })
                )
            );
        }
    }]);

    return MdUnfoldMore;
}(React.Component);

exports.default = MdUnfoldMore;
module.exports = exports['default'];