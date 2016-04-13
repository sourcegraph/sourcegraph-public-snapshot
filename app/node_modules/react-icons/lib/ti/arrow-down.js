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

var TiArrowDown = function (_React$Component) {
    _inherits(TiArrowDown, _React$Component);

    function TiArrowDown() {
        _classCallCheck(this, TiArrowDown);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiArrowDown).apply(this, arguments));
    }

    _createClass(TiArrowDown, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm27.845000000000002 22.155c-0.6499999999999986-0.6499999999999986-1.7049999999999983-0.6499999999999986-2.3566666666666656 0l-3.821666666666669 3.8216666666666654v-12.643333333333333c0-0.9199999999999999-0.745000000000001-1.666666666666666-1.6666666666666679-1.666666666666666s-1.6666666666666679 0.7466666666666661-1.6666666666666679 1.666666666666666v12.643333333333333l-3.821666666666667-3.8216666666666654c-0.6500000000000004-0.6499999999999986-1.705-0.6499999999999986-2.3566666666666674 0s-0.6500000000000004 1.7049999999999983 0 2.3566666666666656l7.845000000000002 7.845000000000002 7.844999999999999-7.844999999999999c0.6499999999999986-0.6499999999999986 0.6499999999999986-1.7049999999999983 0-2.3566666666666656z' })
                )
            );
        }
    }]);

    return TiArrowDown;
}(React.Component);

exports.default = TiArrowDown;
module.exports = exports['default'];