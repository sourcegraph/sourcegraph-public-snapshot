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

var MdFilter8 = function (_React$Component) {
    _inherits(MdFilter8, _React$Component);

    function MdFilter8() {
        _classCallCheck(this, MdFilter8);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdFilter8).apply(this, arguments));
    }

    _createClass(MdFilter8, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm21.64 18.36v3.2833333333333314h3.3599999999999994v-3.2833333333333314h-3.3599999999999994z m0-6.720000000000001v3.360000000000001h3.3599999999999994v-3.3599999999999994h-3.3599999999999994z m0 13.360000000000001q-1.3283333333333331 0-2.3049999999999997-0.9766666666666666t-0.9750000000000014-2.383333333333333v-2.5q0-1.0166666666666657 0.7416666666666671-1.7583333333333329t1.7583333333333329-0.7433333333333323q-1.0166666666666657 0-1.7583333333333329-0.7416666666666671t-0.7416666666666671-1.7599999999999998v-2.5q0-1.4066666666666663 0.9766666666666666-2.3433333333333337t2.306666666666665-0.9383333333333326h3.356666666666669q1.4066666666666663 0 2.383333333333333 0.9383333333333326t0.9766666666666666 2.3433333333333337v2.5q0 1.0166666666666675-0.7416666666666671 1.7583333333333329t-1.7566666666666677 0.7400000000000002q1.0166666666666657 0 1.7583333333333329 0.7416666666666671t0.7433333333333323 1.7566666666666677v2.5q0 1.4066666666666663-1.0166666666666657 2.383333333333333t-2.34333333333333 0.9833333333333307h-3.3583333333333343z m13.36 3.3599999999999994v-23.36h-23.36v23.36h23.36z m0-26.72q1.3283333333333331-4.440892098500626e-16 2.3433333333333337 1.0166666666666662t1.0166666666666657 2.3416666666666663v23.358333333333334q0 1.3283333333333331-1.0166666666666657 2.3049999999999997t-2.3433333333333337 0.9783333333333353h-23.36q-1.3283333333333331 0-2.3049999999999997-0.9766666666666666t-0.9749999999999996-2.3049999999999997v-23.358333333333338q0-1.33 0.9766666666666666-2.345t2.3066666666666666-1.0166666666666666h23.356666666666666z m-30 6.720000000000001v26.64h26.64v3.3599999999999994h-26.64q-1.3283333333333331 0-2.3433333333333333-1.0166666666666657t-1.0166666666666666-2.3416666666666686v-26.638333333333332h3.361666666666667z' })
                )
            );
        }
    }]);

    return MdFilter8;
}(React.Component);

exports.default = MdFilter8;
module.exports = exports['default'];