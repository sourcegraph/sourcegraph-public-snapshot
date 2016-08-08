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

var MdPermScanWifi = function (_React$Component) {
    _inherits(MdPermScanWifi, _React$Component);

    function MdPermScanWifi() {
        _classCallCheck(this, MdPermScanWifi);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdPermScanWifi).apply(this, arguments));
    }

    _createClass(MdPermScanWifi, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm18.36 13.360000000000001h3.2833333333333314v-3.360000000000001h-3.2833333333333314v3.3599999999999994z m3.280000000000001 13.28v-10h-3.2833333333333314v10h3.2833333333333314z m-1.6400000000000006-21.64q10.625 0 20 7.109999999999999l-20 24.53333333333333-20-24.61333333333333q9.216666666666667-7.030000000000001 20-7.030000000000001z' })
                )
            );
        }
    }]);

    return MdPermScanWifi;
}(React.Component);

exports.default = MdPermScanWifi;
module.exports = exports['default'];