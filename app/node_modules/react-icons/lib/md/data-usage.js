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

var MdDataUsage = function (_React$Component) {
    _inherits(MdDataUsage, _React$Component);

    function MdDataUsage() {
        _classCallCheck(this, MdDataUsage);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdDataUsage).apply(this, arguments));
    }

    _createClass(MdDataUsage, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm20 31.640000000000004q5.625 0 9.063333333333333-4.375l4.375 2.576666666666668q-5 6.79666666666667-13.438333333333333 6.79666666666667-6.875 0-11.758333333333333-4.883333333333333t-4.883333333333333-11.75500000000001q-4.440892098500626e-16-6.486666666666666 4.336666666666666-11.213333333333333t10.663333333333334-5.3500000000000005v5.000000000000001q-4.216666666666667 0.6233333333333331-7.109999999999999 3.905000000000001t-2.8916666666666675 7.658333333333331q0 4.841666666666669 3.4000000000000004 8.240000000000002t8.240000000000002 3.3999999999999986z m1.6400000000000006-28.200000000000003q6.328333333333333 0.6233333333333331 10.663333333333334 5.350000000000001t4.336666666666673 11.209999999999997q0 3.75-1.4066666666666663 6.800000000000001l-4.375-2.580000000000002q0.783333333333335-2.1883333333333326 0.783333333333335-4.216666666666669 0-4.376666666666667-2.8916666666666657-7.658333333333333t-7.109999999999999-3.9066666666666663v-5z' })
                )
            );
        }
    }]);

    return MdDataUsage;
}(React.Component);

exports.default = MdDataUsage;
module.exports = exports['default'];