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

var MdWeekend = function (_React$Component) {
    _inherits(MdWeekend, _React$Component);

    function MdWeekend() {
        _classCallCheck(this, MdWeekend);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdWeekend).apply(this, arguments));
    }

    _createClass(MdWeekend, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm30 8.360000000000001q1.3283333333333331 0 2.3433333333333337 0.9766666666666666t1.0166666666666657 2.3049999999999997v3.591666666666667q-1.4866666666666681 0.5500000000000007-2.423333333333332 1.8000000000000007t-0.9366666666666674 2.8866666666666667v3.4383333333333326h-20v-3.4383333333333326q0-1.6400000000000006-0.9399999999999995-2.8900000000000006t-2.421666666666667-1.7966666666666669v-3.5900000000000016q0-1.3283333333333331 1.0166666666666666-2.3049999999999997t2.341666666666666-0.9766666666666666h20z m5 8.28q1.3283333333333331 0 2.3433333333333337 1.0166666666666657t1.0166666666666657 2.3416666666666686v8.358333333333334q0 1.3283333333333331-1.0166666666666657 2.3049999999999997t-2.3433333333333337 0.9783333333333353h-30q-1.3283333333333331 0-2.3433333333333333-0.9766666666666666t-1.0166666666666666-2.3049999999999997v-8.358333333333338q0-1.3299999999999983 1.0166666666666666-2.344999999999999t2.3433333333333333-1.0166666666666657 2.3433333333333337 1.0166666666666657 1.0166666666666657 2.3433333333333337v5h23.28v-5q0-1.3299999999999983 1.0166666666666657-2.344999999999999t2.3416666666666686-1.0166666666666657z' })
                )
            );
        }
    }]);

    return MdWeekend;
}(React.Component);

exports.default = MdWeekend;
module.exports = exports['default'];