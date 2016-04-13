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

var MdLooks3 = function (_React$Component) {
    _inherits(MdLooks3, _React$Component);

    function MdLooks3() {
        _classCallCheck(this, MdLooks3);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdLooks3).apply(this, arguments));
    }

    _createClass(MdLooks3, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm25 17.5v-2.5q0-1.4066666666666663-0.9766666666666666-2.383333333333333t-2.3049999999999997-0.9766666666666666h-6.718333333333334v3.3599999999999994h6.716666666666669v3.3599999999999994h-3.3583333333333343v3.2833333333333314h3.3599999999999994v3.356666666666669h-6.718333333333334v3.361666666666668h6.716666666666669q1.3299999999999983 0 2.306666666666665-0.9766666666666666t0.9766666666666666-2.3850000000000016v-2.5q0-1.0916666666666686-0.7033333333333331-1.7950000000000017t-1.7966666666666669-0.7049999999999983q1.0933333333333337 0 1.7966666666666669-0.6999999999999993t0.7033333333333331-1.8000000000000007z m6.716666666666669-12.5q1.3299999999999983 0 2.306666666666665 1.0166666666666666t0.9766666666666666 2.3400000000000007v23.28333333333333q0 1.326666666666668-0.9766666666666666 2.3416666666666686t-2.306666666666665 1.0183333333333309h-23.35666666666667q-1.3283333333333314 0-2.343333333333332-1.0166666666666657t-1.0166666666666675-2.34333333333333v-23.28333333333334q0-1.3266666666666653 1.0166666666666666-2.341666666666665t2.3400000000000007-1.0150000000000006h23.36z' })
                )
            );
        }
    }]);

    return MdLooks3;
}(React.Component);

exports.default = MdLooks3;
module.exports = exports['default'];