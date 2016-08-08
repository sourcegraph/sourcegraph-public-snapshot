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

var MdFilter9Plus = function (_React$Component) {
    _inherits(MdFilter9Plus, _React$Component);

    function MdFilter9Plus() {
        _classCallCheck(this, MdFilter9Plus);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdFilter9Plus).apply(this, arguments));
    }

    _createClass(MdFilter9Plus, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm35 15v-10h-23.36v23.36h23.36v-10h-3.3599999999999994v3.2833333333333314h-3.2833333333333314v-3.2833333333333314h-3.356666666666669v-3.3599999999999994h3.3583333333333343v-3.3599999999999994h3.2833333333333314v3.3599999999999994h3.3583333333333343z m0-13.360000000000001q1.3283333333333331 0 2.3433333333333337 1.0166666666666666t1.0166666666666657 2.341666666666667v23.358333333333334q0 1.3283333333333331-1.0166666666666657 2.3049999999999997t-2.3433333333333337 0.9783333333333388h-23.36q-1.3283333333333331 0-2.3049999999999997-0.9766666666666666t-0.9749999999999996-2.3049999999999997v-23.358333333333338q0-1.33 0.9766666666666666-2.345t2.3066666666666666-1.0166666666666666h23.356666666666666z m-16.64 13.360000000000001h1.6400000000000006v-1.6400000000000006h-1.6400000000000006v1.6400000000000006z m5 5q0 1.4066666666666663-1.0166666666666657 2.383333333333333t-2.3416666666666686 0.9766666666666666h-5.001666666666665v-3.3599999999999994h5v-1.6400000000000006h-1.6383333333333319q-1.3283333333333331 0-2.3433333333333337-0.9766666666666666t-1.0133333333333336-2.383333333333333v-1.6400000000000006q0-1.4066666666666663 1.0166666666666675-2.383333333333333t2.338333333333331-0.9766666666666666h1.6400000000000006q1.3299999999999983 0 2.344999999999999 0.9766666666666666t1.0166666666666657 2.383333333333333v6.640000000000001z m-18.36-11.639999999999999v26.64h26.64v3.3599999999999994h-26.64q-1.3283333333333331 0-2.3433333333333333-1.0166666666666657t-1.0166666666666666-2.3416666666666686v-26.638333333333332h3.361666666666667z' })
                )
            );
        }
    }]);

    return MdFilter9Plus;
}(React.Component);

exports.default = MdFilter9Plus;
module.exports = exports['default'];