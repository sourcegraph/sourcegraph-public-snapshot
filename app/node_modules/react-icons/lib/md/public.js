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

var MdPublic = function (_React$Component) {
    _inherits(MdPublic, _React$Component);

    function MdPublic() {
        _classCallCheck(this, MdPublic);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdPublic).apply(this, arguments));
    }

    _createClass(MdPublic, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm29.843333333333334 28.983333333333334q3.5166666666666657-3.75 3.5166666666666657-8.983333333333334 0-4.140000000000001-2.306666666666665-7.5t-6.053333333333335-4.843333333333333v0.703333333333334q0 1.3283333333333331-1.0166666666666657 2.3049999999999997t-2.3433333333333337 0.9749999999999996h-3.2833333333333314v3.3599999999999994q0 0.7050000000000001-0.5066666666666677 1.173333333333332t-1.211666666666666 0.466666666666665h-3.283333333333333v3.361666666666668h10.000000000000002q0.7049999999999983 0 1.173333333333332 0.46999999999999886t0.466666666666665 1.1716666666666669v5h1.6416666666666657q2.3433333333333337 0 3.203333333333333 2.3433333333333337z m-11.483333333333334 4.219999999999999v-3.203333333333333q-1.3283333333333331 0-2.3433333333333337-1.0166666666666657t-1.0166666666666657-2.34v-1.6433333333333344l-7.966666666666668-7.966666666666669q-0.3899999999999997 1.5633333333333326-0.3899999999999997 2.9666666666666686 0 5.079999999999998 3.4000000000000004 8.829999999999998t8.316666666666666 4.375z m1.6400000000000006-29.843333333333334q6.875 8.881784197001252e-16 11.758333333333333 4.883333333333335t4.883333333333333 11.756666666666666-4.883333333333333 11.759999999999998-11.758333333333333 4.88333333333334-11.758333333333333-4.883333333333333-4.883333333333333-11.760000000000005 4.883333333333333-11.756666666666668 11.758333333333333-4.883333333333332z' })
                )
            );
        }
    }]);

    return MdPublic;
}(React.Component);

exports.default = MdPublic;
module.exports = exports['default'];