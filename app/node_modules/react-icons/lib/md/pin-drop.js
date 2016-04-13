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

var MdPinDrop = function (_React$Component) {
    _inherits(MdPinDrop, _React$Component);

    function MdPinDrop() {
        _classCallCheck(this, MdPinDrop);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdPinDrop).apply(this, arguments));
    }

    _createClass(MdPinDrop, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm8.360000000000001 33.36h23.28333333333334v3.2833333333333314h-23.285000000000004v-3.2833333333333314z m8.28-20q0 1.3283333333333331 1.0166666666666657 2.3049999999999997t2.3416666666666686 0.9750000000000014q1.4050000000000011 0 2.383333333333333-0.9766666666666666t0.9750000000000014-2.3066666666666666-1.0166666666666657-2.3433333333333337-2.3400000000000034-1.0133333333333336-2.3433333333333337 1.0166666666666657-1.0166666666666657 2.3433333333333337z m13.36 0q0 3.3599999999999994-2.5 7.93t-5 7.461666666666666l-2.5 2.8916666666666693q-1.0933333333333337-1.1716666666666669-2.7733333333333334-3.203333333333333t-4.453333333333333-6.875-2.7733333333333334-8.205q0-4.140000000000001 2.9299999999999997-7.07t7.07-2.9300000000000006 7.07 2.9300000000000006 2.9299999999999997 7.07z' })
                )
            );
        }
    }]);

    return MdPinDrop;
}(React.Component);

exports.default = MdPinDrop;
module.exports = exports['default'];