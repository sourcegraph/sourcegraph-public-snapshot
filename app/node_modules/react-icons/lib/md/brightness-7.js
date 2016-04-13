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

var MdBrightness7 = function (_React$Component) {
    _inherits(MdBrightness7, _React$Component);

    function MdBrightness7() {
        _classCallCheck(this, MdBrightness7);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdBrightness7).apply(this, arguments));
    }

    _createClass(MdBrightness7, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm20 13.360000000000001q2.7333333333333343 0 4.688333333333333 1.9533333333333331t1.9533333333333331 4.686666666666666-1.9533333333333331 4.690000000000001-4.688333333333333 1.9533333333333331-4.688333333333333-1.9533333333333331-1.9533333333333331-4.690000000000001 1.9533333333333314-4.683333333333334 4.688333333333334-1.9566666666666652z m0 16.64q4.140000000000001 0 7.07-2.9299999999999997t2.9299999999999997-7.07-2.9299999999999997-7.07-7.07-2.9299999999999997-7.07 2.9299999999999997-2.9299999999999997 7.07 2.9299999999999997 7.07 7.07 2.9299999999999997z m13.36-15.546666666666667l5.466666666666669 5.546666666666667-5.466666666666669 5.546666666666667v7.813333333333333h-7.813333333333333l-5.546666666666667 5.466666666666669-5.546666666666667-5.466666666666669h-7.813333333333333v-7.813333333333333l-5.466666666666667-5.546666666666667 5.466666666666667-5.546666666666667v-7.813333333333333h7.813333333333333l5.546666666666667-5.466666666666667 5.546666666666667 5.466666666666667h7.813333333333333v7.813333333333333z' })
                )
            );
        }
    }]);

    return MdBrightness7;
}(React.Component);

exports.default = MdBrightness7;
module.exports = exports['default'];