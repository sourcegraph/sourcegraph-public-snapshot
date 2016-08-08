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

var MdHttps = function (_React$Component) {
    _inherits(MdHttps, _React$Component);

    function MdHttps() {
        _classCallCheck(this, MdHttps);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdHttps).apply(this, arguments));
    }

    _createClass(MdHttps, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm25.156666666666666 13.360000000000001v-3.360000000000001q0-2.1100000000000003-1.5233333333333334-3.6333333333333337t-3.633333333333333-1.5233333333333325-3.633333333333333 1.5233333333333334-1.5233333333333317 3.633333333333333v3.3599999999999994h10.31333333333333z m-5.156666666666666 14.999999999999998q1.3283333333333331 0 2.3433333333333337-1.0166666666666657t1.0166666666666657-2.3416666666666686-1.0166666666666657-2.3416666666666686-2.3433333333333337-1.0166666666666657-2.3433333333333337 1.0166666666666657-1.0166666666666657 2.3400000000000034 1.0166666666666657 2.344999999999999 2.3433333333333337 1.0166666666666657z m10-15q1.3283333333333331 0 2.3433333333333337 0.9766666666666666t1.0166666666666657 2.3049999999999997v16.71666666666667q0 1.3299999999999983-1.0166666666666657 2.306666666666665t-2.3433333333333337 0.9750000000000085h-20q-1.3283333333333331 0-2.3433333333333337-0.9766666666666666t-1.0166666666666666-2.306666666666665v-16.713333333333342q0-1.33 1.0166666666666666-2.3066666666666666t2.3433333333333337-0.9766666666666666h1.6400000000000006v-3.360000000000001q0-3.4383333333333344 2.461666666666666-5.9t5.898333333333333-2.458333333333333 5.899999999999999 2.461666666666667 2.4583333333333357 5.8966666666666665v3.3599999999999994h1.6416666666666657z' })
                )
            );
        }
    }]);

    return MdHttps;
}(React.Component);

exports.default = MdHttps;
module.exports = exports['default'];