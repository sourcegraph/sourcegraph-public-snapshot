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

var MdSwitchVideo = function (_React$Component) {
    _inherits(MdSwitchVideo, _React$Component);

    function MdSwitchVideo() {
        _classCallCheck(this, MdSwitchVideo);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdSwitchVideo).apply(this, arguments));
    }

    _createClass(MdSwitchVideo, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm21.64 25.86l5.859999999999999-5.859999999999999-5.859999999999999-5.859999999999999v4.216666666666669h-10v-4.216666666666667l-5.783333333333334 5.859999999999998 5.783333333333334 5.859999999999999v-4.216666666666669h10v4.216666666666669z m8.36-10l6.640000000000001-6.716666666666669v21.716666666666665l-6.640000000000001-6.7166666666666615v5.856666666666666q0 0.7049999999999983-0.466666666666665 1.173333333333332t-1.173333333333332 0.466666666666665h-23.360000000000003q-0.7033333333333331 0-1.1716666666666669-0.466666666666665t-0.46666666666666634-1.173333333333332v-20q0-0.6999999999999993 0.4666666666666668-1.17t1.1716666666666664-0.4666666666666668h23.36q0.7033333333333331 0 1.1716666666666669 0.4666666666666668t0.4683333333333337 1.17v5.861666666666666z' })
                )
            );
        }
    }]);

    return MdSwitchVideo;
}(React.Component);

exports.default = MdSwitchVideo;
module.exports = exports['default'];