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

var TiArrowUp = function (_React$Component) {
    _inherits(TiArrowUp, _React$Component);

    function TiArrowUp() {
        _classCallCheck(this, TiArrowUp);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiArrowUp).apply(this, arguments));
    }

    _createClass(TiArrowUp, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm21.666666666666668 9.31l-7.844999999999999 7.845000000000001c-0.6500000000000004 0.6499999999999986-0.6500000000000004 1.7049999999999983 0 2.3566666666666656s1.705 0.6499999999999986 2.3566666666666656 0l3.8216666666666654-3.821666666666667v12.643333333333336c0 0.9200000000000017 0.745000000000001 1.6666666666666679 1.6666666666666679 1.6666666666666679s1.6666666666666679-0.7466666666666661 1.6666666666666679-1.6666666666666679v-12.643333333333336l3.8216666666666654 3.821666666666667c0.3249999999999993 0.3249999999999993 0.75 0.4883333333333333 1.1783333333333346 0.4883333333333333s0.8533333333333317-0.163333333333334 1.1783333333333346-0.4883333333333333c0.6499999999999986-0.6499999999999986 0.6499999999999986-1.7049999999999983 0-2.3566666666666656l-7.845000000000002-7.845000000000001z' })
                )
            );
        }
    }]);

    return TiArrowUp;
}(React.Component);

exports.default = TiArrowUp;
module.exports = exports['default'];