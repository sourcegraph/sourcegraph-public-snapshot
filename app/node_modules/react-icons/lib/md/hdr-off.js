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

var MdHdrOff = function (_React$Component) {
    _inherits(MdHdrOff, _React$Component);

    function MdHdrOff() {
        _classCallCheck(this, MdHdrOff);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdHdrOff).apply(this, arguments));
    }

    _createClass(MdHdrOff, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm4.140000000000001 4.140000000000001q7.656666666666666 7.733333333333334 31.875 31.71666666666667l-1.8766666666666652 1.7999999999999972-12.656666666666666-12.658333333333331h-5.623333333333335v-5.704999999999998l-2.5-2.5v8.206666666666663h-2.5v-4.141666666666666h-3.3583333333333343v4.141666666666666h-2.5v-10h2.5v3.3583333333333343h3.3583333333333343v-3.3583333333333343h0.625l-9.14-9.143333333333334z m17.5 13.36h-0.625l-2.498333333333335-2.5h3.123333333333335q1.0166666666666657 0 1.7583333333333329 0.7416666666666671t0.7399999999999984 1.7583333333333329v3.203333333333333l-2.5-2.5v-0.7033333333333331z m7.5 0v1.6400000000000006h3.3599999999999994v-1.6400000000000006h-3.3599999999999994z m0 7.5h-0.625l-1.8766666666666652-1.7966666666666669v-8.203333333333333h5.861666666666665q1.0166666666666657 0 1.7583333333333329 0.7416666666666671t0.7416666666666671 1.7583333333333329v1.6400000000000006q0 1.7966666666666669-1.4833333333333343 2.3433333333333337l1.4833333333333343 3.5166666666666657h-2.5l-1.4833333333333343-3.361666666666668h-1.8766666666666652v3.361666666666668z' })
                )
            );
        }
    }]);

    return MdHdrOff;
}(React.Component);

exports.default = MdHdrOff;
module.exports = exports['default'];