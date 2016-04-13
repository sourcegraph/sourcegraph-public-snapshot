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

var MdHeadsetMic = function (_React$Component) {
    _inherits(MdHeadsetMic, _React$Component);

    function MdHeadsetMic() {
        _classCallCheck(this, MdHeadsetMic);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdHeadsetMic).apply(this, arguments));
    }

    _createClass(MdHeadsetMic, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm20 1.6400000000000001q6.25 0 10.625 4.413333333333334t4.375 10.586666666666666v16.71666666666667q0 2.0333333333333314-1.4833333333333343 3.5166666666666657t-3.5166666666666657 1.4833333333333343h-10v-3.356666666666669h11.64v-1.6416666666666657h-6.640000000000001v-13.358333333333334h6.640000000000001v-3.3633333333333333q0-4.843333333333334-3.3999999999999986-8.241666666666667t-8.240000000000002-3.3983333333333325-8.241666666666667 3.4033333333333333-3.4000000000000004 8.241666666666665v3.3583333333333343h6.6416666666666675v13.358333333333334h-5q-2.033333333333333 0-3.5166666666666666-1.4833333333333343t-1.4833333333333334-3.5183333333333344v-11.716666666666665q0-6.173333333333334 4.375-10.59t10.625-4.41z' })
                )
            );
        }
    }]);

    return MdHeadsetMic;
}(React.Component);

exports.default = MdHeadsetMic;
module.exports = exports['default'];