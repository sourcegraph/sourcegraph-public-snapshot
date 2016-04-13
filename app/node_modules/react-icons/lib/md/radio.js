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

var MdRadio = function (_React$Component) {
    _inherits(MdRadio, _React$Component);

    function MdRadio() {
        _classCallCheck(this, MdRadio);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdRadio).apply(this, arguments));
    }

    _createClass(MdRadio, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm33.36 20v-6.639999999999999h-26.716666666666665v6.639999999999999h20v-3.3599999999999994h3.3566666666666656v3.3599999999999994h3.3616666666666646z m-21.72 13.36q2.033333333333333 0 3.5166666666666657-1.4833333333333343t1.4833333333333343-3.5166666666666657-1.4833333333333343-3.5166666666666657-3.5166666666666657-1.4833333333333343-3.5166666666666657 1.4833333333333343-1.4833333333333334 3.5166666666666657 1.4833333333333334 3.5166666666666657 3.5166666666666657 1.4833333333333343z m-6.25-23.126666666666665l21.093333333333334-8.590000000000002 1.0933333333333337 2.8133333333333344-13.750000000000002 5.543333333333333h19.53333333333333q1.4050000000000011 0 2.3416666666666686 0.9783333333333335t0.9399999999999977 2.383333333333333v20q0 1.3283333333333331-0.9383333333333326 2.3049999999999997t-2.3416666666666686 0.9766666666666666h-26.72q-1.4083333333333323 0-2.344999999999999-0.9766666666666666t-0.9383333333333335-2.3049999999999997v-20q0-2.3433333333333337 2.033333333333333-3.125z' })
                )
            );
        }
    }]);

    return MdRadio;
}(React.Component);

exports.default = MdRadio;
module.exports = exports['default'];