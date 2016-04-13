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

var MdHdrStrong = function (_React$Component) {
    _inherits(MdHdrStrong, _React$Component);

    function MdHdrStrong() {
        _classCallCheck(this, MdHdrStrong);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdHdrStrong).apply(this, arguments));
    }

    _createClass(MdHdrStrong, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm8.360000000000001 23.36q1.3283333333333331 0 2.3049999999999997-1.0166666666666657t0.9749999999999996-2.3416666666666686-0.9766666666666666-2.3416666666666686-2.3066666666666666-1.0166666666666657-2.3433333333333337 1.0166666666666657-1.0133333333333336 2.3400000000000034 1.0166666666666666 2.344999999999999 2.343333333333333 1.0166666666666657z m0-10q2.7333333333333325 0 4.688333333333334 1.9533333333333331t1.9516666666666644 4.686666666666667-1.9499999999999993 4.690000000000001-4.690000000000001 1.9533333333333331-4.726666666666667-1.9533333333333331-1.9916666666666663-4.690000000000001 1.9916666666666671-4.683333333333334 4.726666666666666-1.9533333333333331z m20-3.3599999999999994q4.140000000000001 0 7.07 2.9299999999999997t2.9299999999999997 7.07-2.9299999999999997 7.07-7.07 2.9299999999999997-7.07-2.9299999999999997-2.9299999999999997-7.07 2.9299999999999997-7.07 7.07-2.9299999999999997z' })
                )
            );
        }
    }]);

    return MdHdrStrong;
}(React.Component);

exports.default = MdHdrStrong;
module.exports = exports['default'];