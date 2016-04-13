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

var MdGpsFixed = function (_React$Component) {
    _inherits(MdGpsFixed, _React$Component);

    function MdGpsFixed() {
        _classCallCheck(this, MdGpsFixed);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdGpsFixed).apply(this, arguments));
    }

    _createClass(MdGpsFixed, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm20 31.640000000000004q4.843333333333334 0 8.241666666666667-3.400000000000002t3.3999999999999986-8.240000000000002-3.3999999999999986-8.241666666666667-8.241666666666667-3.4000000000000004-8.241666666666667 3.4000000000000004-3.4000000000000004 8.241666666666667 3.4000000000000004 8.241666666666667 8.241666666666667 3.3999999999999986z m14.921666666666667-13.280000000000001h3.4383333333333326v3.283333333333335h-3.4383333333333326q-0.5466666666666669 5.233333333333334-4.296666666666667 8.983333333333334t-8.983333333333334 4.296666666666667v3.4383333333333326h-3.2833333333333314v-3.4383333333333326q-5.233333333333334-0.5466666666666669-8.983333333333333-4.296666666666667t-4.300000000000001-8.983333333333334h-3.435000000000001v-3.2833333333333314h3.438333333333334q0.5466666666666669-5.233333333333334 4.296666666666666-8.983333333333333t8.983333333333334-4.300000000000001v-3.4350000000000067h3.2833333333333314v3.4383333333333335q5.233333333333334 0.5466666666666669 8.983333333333334 4.296666666666667t4.299999999999997 8.983333333333333z m-14.921666666666667-5q2.7333333333333343 0 4.688333333333333 1.9533333333333331t1.9533333333333331 4.686666666666664-1.9533333333333331 4.690000000000001-4.688333333333333 1.9533333333333331-4.688333333333333-1.9533333333333331-1.9533333333333331-4.690000000000001 1.9533333333333314-4.683333333333334 4.688333333333334-1.9566666666666652z' })
                )
            );
        }
    }]);

    return MdGpsFixed;
}(React.Component);

exports.default = MdGpsFixed;
module.exports = exports['default'];