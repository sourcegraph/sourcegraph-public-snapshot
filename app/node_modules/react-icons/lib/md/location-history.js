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

var MdLocationHistory = function (_React$Component) {
    _inherits(MdLocationHistory, _React$Component);

    function MdLocationHistory() {
        _classCallCheck(this, MdLocationHistory);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdLocationHistory).apply(this, arguments));
    }

    _createClass(MdLocationHistory, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm30 26.64v-1.4833333333333343q0-2.2666666666666657-3.4383333333333326-3.7133333333333347t-6.561666666666667-1.4433333333333316-6.566666666666666 1.4433333333333351-3.4333333333333336 3.711666666666666v1.4833333333333343h20z m-10-17.811666666666667q-1.875 0-3.203333333333333 1.3283333333333331t-1.33 3.203333333333335 1.33 3.163333333333332 3.203333333333333 1.288333333333334 3.203333333333333-1.288333333333334 1.3283333333333331-3.163333333333334-1.3299999999999983-3.203333333333333-3.201666666666668-1.3283333333333331z m11.64-5.466666666666667q1.3283333333333331 0 2.3433333333333337 0.9749999999999996t1.0166666666666657 2.3050000000000006v23.358333333333334q0 1.3299999999999983-1.0166666666666657 2.344999999999999t-2.3433333333333337 1.0166666666666657h-6.640000000000001l-5 5-5-5h-6.639999999999999q-1.4066666666666663 0-2.383333333333333-1.0166666666666657t-0.9766666666666683-2.344999999999999v-23.356666666666666q0-1.328333333333334 0.9766666666666666-2.3050000000000015t2.383333333333333-0.9766666666666666h23.283333333333335z' })
                )
            );
        }
    }]);

    return MdLocationHistory;
}(React.Component);

exports.default = MdLocationHistory;
module.exports = exports['default'];