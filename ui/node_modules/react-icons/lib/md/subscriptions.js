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

var MdSubscriptions = function (_React$Component) {
    _inherits(MdSubscriptions, _React$Component);

    function MdSubscriptions() {
        _classCallCheck(this, MdSubscriptions);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdSubscriptions).apply(this, arguments));
    }

    _createClass(MdSubscriptions, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm26.64 26.64l-10-5.390000000000001v10.86z m10-6.640000000000001v13.36q0 1.3283333333333331-0.9766666666666666 2.3049999999999997t-2.3049999999999997 0.9750000000000014h-26.71666666666667q-1.3299999999999992 0-2.3066666666666658-0.9766666666666666t-0.9750000000000001-2.306666666666665v-13.35666666666667q0-1.3283333333333331 0.976666666666667-2.3433333333333337t2.3066666666666666-1.0166666666666657h26.716666666666665q1.3299999999999983 0 2.306666666666665 1.0166666666666657t0.9733333333333434 2.3433333333333337z m-6.640000000000001-16.64v3.283333333333334h-20v-3.283333333333333h20z m3.3599999999999994 10h-26.716666666666665v-3.3599999999999994h26.716666666666665v3.3599999999999994z' })
                )
            );
        }
    }]);

    return MdSubscriptions;
}(React.Component);

exports.default = MdSubscriptions;
module.exports = exports['default'];