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

var MdRvHookup = function (_React$Component) {
    _inherits(MdRvHookup, _React$Component);

    function MdRvHookup() {
        _classCallCheck(this, MdRvHookup);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdRvHookup).apply(this, arguments));
    }

    _createClass(MdRvHookup, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm28.36 3.3600000000000003l5 4.999999999999999-5 5v-3.3599999999999994h-13.36v-3.3599999999999994h13.36v-3.283333333333333z m1.6400000000000006 20v-5h-6.640000000000001v5h6.640000000000001z m-11.64 10q0.7033333333333331 0 1.1716666666666669-0.5083333333333329t0.466666666666665-1.2100000000000009-0.466666666666665-1.1716666666666669-1.173333333333332-0.466666666666665-1.211666666666666 0.466666666666665-0.5100000000000016 1.173333333333332 0.5083333333333329 1.211666666666666 1.2100000000000009 0.509999999999998z m15-5h3.2833333333333314v3.2833333333333314h-13.283333333333331q0 2.030000000000001-1.4833333333333343 3.5133333333333354t-3.5166666666666657 1.4833333333333343-3.5166666666666657-1.4833333333333343-1.4833333333333343-3.5166666666666657h-3.3599999999999994q-1.3283333333333331 0-2.3433333333333337-0.9750000000000014t-1.0166666666666666-2.306666666666665v-5h11.719999999999999v-5h-6.716666666666669v3.2833333333333314l-5-5 5-5v3.3583333333333343h18.35666666666667q1.3299999999999983 0 2.344999999999999 1.0166666666666657t1.0166666666666657 2.3433333333333337v10z' })
                )
            );
        }
    }]);

    return MdRvHookup;
}(React.Component);

exports.default = MdRvHookup;
module.exports = exports['default'];