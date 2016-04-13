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

var MdPageview = function (_React$Component) {
    _inherits(MdPageview, _React$Component);

    function MdPageview() {
        _classCallCheck(this, MdPageview);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdPageview).apply(this, arguments));
    }

    _createClass(MdPageview, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm27.96666666666667 30.313333333333333l2.344999999999999-2.3433333333333337-4.843333333333334-4.843333333333334q1.173333333333332-1.9533333333333331 1.173333333333332-3.9833333333333343 0-3.126666666666665-2.1883333333333326-5.316666666666666t-5.311666666666667-2.1866666666666674-5.313333333333333 2.1883333333333326-2.1883333333333326 5.313333333333333 2.1883333333333326 5.313333333333333 5.313333333333333 2.1883333333333326q2.0333333333333314 0 3.9833333333333343-1.1716666666666669z m5.393333333333331-23.673333333333336q1.3283333333333331 0 2.3049999999999997 1.0166666666666666t0.9750000000000014 2.3416666666666677v20q0 1.326666666666668-0.9766666666666666 2.341666666666665t-2.306666666666665 1.0166666666666657h-26.713333333333335q-1.330000000000001 0-2.3066666666666675-1.0166666666666657t-0.9766666666666666-2.3399999999999963v-20q0-1.33 0.9766666666666666-2.3450000000000006t2.3050000000000006-1.0166666666666666h26.71666666666667z m-14.219999999999999 8.360000000000003q1.7166666666666686 0 2.9666666666666686 1.211666666666666t1.25 2.9299999999999997-1.25 2.9666666666666686-2.9666666666666686 1.25-2.9299999999999997-1.25-1.2100000000000009-2.965 1.2100000000000009-2.9299999999999997 2.9299999999999997-1.2133333333333347z' })
                )
            );
        }
    }]);

    return MdPageview;
}(React.Component);

exports.default = MdPageview;
module.exports = exports['default'];