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

var MdCameraRoll = function (_React$Component) {
    _inherits(MdCameraRoll, _React$Component);

    function MdCameraRoll() {
        _classCallCheck(this, MdCameraRoll);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdCameraRoll).apply(this, arguments));
    }

    _createClass(MdCameraRoll, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm33.36 15v-3.3599999999999994h-3.3599999999999994v3.3599999999999994h3.3599999999999994z m0 15v-3.3599999999999994h-3.3599999999999994v3.3599999999999994h3.3599999999999994z m-6.719999999999999-15v-3.3599999999999994h-3.2833333333333314v3.3599999999999994h3.2833333333333314z m0 15v-3.3599999999999994h-3.2833333333333314v3.3599999999999994h3.2833333333333314z m-6.640000000000001-15v-3.3599999999999994h-3.3599999999999994v3.3599999999999994h3.3599999999999994z m0 15v-3.3599999999999994h-3.3599999999999994v3.3599999999999994h3.3599999999999994z m3.3599999999999994-21.64h13.283333333333331v25h-13.283333333333331q0 1.3283333333333331-1.0166666666666657 2.3049999999999997t-2.3416666666666686 0.9750000000000014h-13.35833333333333q-1.328333333333334 0-2.3050000000000015-0.9766666666666666t-0.9766666666666666-2.306666666666665v-25q0-1.3283333333333331 0.9766666666666666-2.3433333333333337t2.3033333333333337-1.0133333333333354h1.7166666666666677v-1.6383333333333336q0-0.7033333333333331 0.47000000000000064-1.211666666666667t1.171666666666665-0.5083333333333329h6.640000000000001q0.7033333333333331 0 1.211666666666666 0.5083333333333333t0.5100000000000016 1.2116666666666664v1.6383333333333336h1.6383333333333319q1.3283333333333331 0 2.3433333333333337 1.0166666666666666t1.0166666666666657 2.343333333333333z' })
                )
            );
        }
    }]);

    return MdCameraRoll;
}(React.Component);

exports.default = MdCameraRoll;
module.exports = exports['default'];