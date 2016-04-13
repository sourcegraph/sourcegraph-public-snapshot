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

var MdFilter3 = function (_React$Component) {
    _inherits(MdFilter3, _React$Component);

    function MdFilter3() {
        _classCallCheck(this, MdFilter3);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdFilter3).apply(this, arguments));
    }

    _createClass(MdFilter3, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm28.36 21.64q0 1.4066666666666663-1.0166666666666657 2.383333333333333t-2.34 0.9766666666666666h-6.638333333333332v-3.3599999999999994h6.634999999999998v-3.2833333333333314h-3.3583333333333343v-3.356666666666669h3.3583333333333343v-3.3633333333333333h-6.638333333333335v-3.283333333333333h6.638333333333335q1.408333333333335 0 2.383333333333333 0.9399999999999995t0.9783333333333317 2.3433333333333337v2.5q0 1.0166666666666657-0.7416666666666671 1.7583333333333329t-1.7583333333333329 0.7416666666666671q1.0166666666666657 0 1.7583333333333329 0.7416666666666671t0.7416666666666671 1.7583333333333329v2.5z m-23.36-13.28v26.64h26.64v3.3599999999999994h-26.64q-1.3283333333333331 0-2.3433333333333333-1.0166666666666657t-1.0166666666666666-2.3416666666666686v-26.638333333333332h3.361666666666667z m30 20v-23.36h-23.36v23.36h23.36z m0-26.72q1.3283333333333331-4.440892098500626e-16 2.3433333333333337 1.0166666666666662t1.0166666666666657 2.3416666666666663v23.358333333333334q0 1.3283333333333331-1.0166666666666657 2.3049999999999997t-2.3433333333333337 0.9783333333333353h-23.36q-1.3283333333333331 0-2.3049999999999997-0.9766666666666666t-0.9749999999999996-2.3049999999999997v-23.358333333333338q0-1.33 0.9766666666666666-2.345t2.3066666666666666-1.0166666666666666h23.356666666666666z' })
                )
            );
        }
    }]);

    return MdFilter3;
}(React.Component);

exports.default = MdFilter3;
module.exports = exports['default'];