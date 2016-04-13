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

var MdCallSplit = function (_React$Component) {
    _inherits(MdCallSplit, _React$Component);

    function MdCallSplit() {
        _classCallCheck(this, MdCallSplit);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdCallSplit).apply(this, arguments));
    }

    _createClass(MdCallSplit, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm16.64 6.640000000000001l-3.828333333333333 3.826666666666666 8.828333333333333 8.83v14.063333333333333h-3.2833333333333314v-12.656666666666666l-7.8866666666666685-7.890000000000001-3.828333333333334 3.828333333333333v-10h10z m6.719999999999999 0h10v10l-3.826666666666668-3.828333333333333-4.844999999999999 4.843333333333334-2.3433333333333337-2.3433333333333337 4.843333333333334-4.843333333333334z' })
                )
            );
        }
    }]);

    return MdCallSplit;
}(React.Component);

exports.default = MdCallSplit;
module.exports = exports['default'];