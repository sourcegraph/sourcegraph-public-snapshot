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

var MdSpeaker = function (_React$Component) {
    _inherits(MdSpeaker, _React$Component);

    function MdSpeaker() {
        _classCallCheck(this, MdSpeaker);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdSpeaker).apply(this, arguments));
    }

    _createClass(MdSpeaker, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm20 20q2.0333333333333314 0 3.5166666666666657 1.4833333333333343t1.4833333333333343 3.5166666666666657-1.4833333333333343 3.5166666666666657-3.5166666666666657 1.4833333333333343-3.5166666666666657-1.4833333333333343-1.4833333333333343-3.5166666666666657 1.4833333333333343-3.5166666666666657 3.5166666666666657-1.4833333333333343z m0 13.36q3.4383333333333326 0 5.899999999999999-2.461666666666666t2.4583333333333357-5.898333333333333-2.461666666666666-5.899999999999999-5.896666666666668-2.4583333333333357-5.899999999999999 2.4583333333333357-2.458333333333334 5.899999999999999 2.461666666666668 5.899999999999999 5.896666666666665 2.4583333333333357z m0-26.720000000000002q-1.4066666666666663 0-2.383333333333333 1.0166666666666666t-0.9766666666666666 2.3416666666666677 0.9766666666666666 2.341666666666667 2.383333333333333 1.0166666666666675q1.3283333333333331 0 2.3433333333333337-1.0166666666666657t1.0166666666666657-2.34-1.0166666666666657-2.3450000000000006-2.3433333333333337-1.0166666666666657z m8.36-3.2800000000000002q1.3283333333333331 0 2.3049999999999997 0.9766666666666666t0.9750000000000014 2.3050000000000006v26.71666666666667q0 1.3299999999999983-0.9766666666666666 2.306666666666665t-2.306666666666665 0.9750000000000014h-16.71333333333334q-1.3299999999999983 0-2.306666666666665-0.9766666666666666t-0.9766666666666666-2.306666666666665v-26.713333333333335q0-1.330000000000001 0.9766666666666666-2.3066666666666675t2.3049999999999997-0.9766666666666666h16.716666666666665z' })
                )
            );
        }
    }]);

    return MdSpeaker;
}(React.Component);

exports.default = MdSpeaker;
module.exports = exports['default'];