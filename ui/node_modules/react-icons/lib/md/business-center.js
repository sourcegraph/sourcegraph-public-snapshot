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

var MdBusinessCenter = function (_React$Component) {
    _inherits(MdBusinessCenter, _React$Component);

    function MdBusinessCenter() {
        _classCallCheck(this, MdBusinessCenter);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdBusinessCenter).apply(this, arguments));
    }

    _createClass(MdBusinessCenter, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm23.36 11.64v-3.283333333333333h-6.716666666666669v3.283333333333333h6.716666666666669z m10 0q1.3283333333333331 0 2.3049999999999997 1.0166666666666657t0.9750000000000014 2.341666666666667v5.000000000000002q0 1.326666666666668-0.9766666666666666 2.3416666666666686t-2.306666666666665 1.0166666666666657h-10v-3.361666666666668h-6.716666666666669v3.3599999999999994h-10q-1.4083333333333332 0-2.3450000000000006-0.9766666666666666t-0.94-2.383333333333333v-5q0-1.3283333333333331 0.976666666666667-2.3433333333333337t2.3049999999999997-1.0166666666666657h6.6466666666666665v-3.276666666666669l3.3583333333333325-3.3600000000000003h6.640000000000001l3.3583333333333343 3.3600000000000003v3.283333333333333h6.716666666666669z m-16.720000000000002 15h6.716666666666669v-1.6400000000000006h11.643333333333334v6.640000000000001q0 1.4066666666666663-0.9783333333333317 2.383333333333333t-2.383333333333333 0.9766666666666666h-23.283333333333335q-1.4050000000000002 0-2.383333333333333-0.9766666666666666t-0.9716666666666676-2.3833333333333293v-6.640000000000004h11.64v1.6400000000000006z' })
                )
            );
        }
    }]);

    return MdBusinessCenter;
}(React.Component);

exports.default = MdBusinessCenter;
module.exports = exports['default'];