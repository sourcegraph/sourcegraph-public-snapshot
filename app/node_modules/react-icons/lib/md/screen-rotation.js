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

var MdScreenRotation = function (_React$Component) {
    _inherits(MdScreenRotation, _React$Component);

    function MdScreenRotation() {
        _classCallCheck(this, MdScreenRotation);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdScreenRotation).apply(this, arguments));
    }

    _createClass(MdScreenRotation, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm12.5 35.78333333333333l2.2666666666666675-2.1899999999999977 6.326666666666666 6.328333333333333-1.0933333333333337 0.07833333333333314q-7.813333333333333 0-13.555-5.313333333333333t-6.366666666666666-13.046666666666667h2.5q0.4666666666666668 4.688333333333333 3.125 8.438333333333333t6.796666666666666 5.704999999999998z m12.188333333333333-0.46999999999999886l10.625-10.625-20-20-10.625 10.625z m-7.654999999999998-32.42333333333333l20.075 20.076666666666664q0.7833333333333314 0.7049999999999983 0.7833333333333314 1.7199999999999989t-0.7833333333333314 1.7966666666666669l-10.625 10.625000000000004q-0.6999999999999993 0.7833333333333314-1.7166666666666686 0.7833333333333314t-1.7966666666666669-0.7833333333333314l-20.078333333333333-20.078333333333333q-0.7833333333333319-0.7033333333333331-0.7833333333333319-1.7166666666666668t0.7833333333333332-1.8000000000000007l10.624999999999998-10.621666666666666q0.7033333333333331-0.7833333333333337 1.7166666666666668-0.7833333333333337t1.799999999999999 0.7833333333333332z m10.466666666666665 1.3283333333333331l-2.2666666666666657 2.1900000000000004-6.324999999999999-6.330000000000001 1.091666666666665-0.07833333333333314q7.813333333333333 0 13.555 5.3133333333333335t6.366666666666667 13.046666666666667h-2.5q-0.46666666666666856-4.688333333333333-3.125-8.438333333333333t-6.796666666666667-5.705z' })
                )
            );
        }
    }]);

    return MdScreenRotation;
}(React.Component);

exports.default = MdScreenRotation;
module.exports = exports['default'];