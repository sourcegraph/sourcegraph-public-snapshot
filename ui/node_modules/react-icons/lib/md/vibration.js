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

var MdVibration = function (_React$Component) {
    _inherits(MdVibration, _React$Component);

    function MdVibration() {
        _classCallCheck(this, MdVibration);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdVibration).apply(this, arguments));
    }

    _createClass(MdVibration, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm26.64 31.640000000000004v-23.28333333333334h-13.283333333333333v23.283333333333335h13.283333333333333z m0.8599999999999994-26.640000000000004q1.0933333333333337 0 1.7966666666666669 0.7033333333333331t0.7033333333333331 1.7966666666666669v25q0 1.0933333333333337-0.7033333333333331 1.7966666666666669t-1.7966666666666669 0.7033333333333331h-15q-1.0933333333333337 0-1.7966666666666669-0.7033333333333331t-0.7033333333333331-1.7966666666666669v-25q0-1.0933333333333337 0.7033333333333331-1.7966666666666669t1.7966666666666669-0.7033333333333331h15z m4.140000000000001 23.36v-16.71666666666667h3.3599999999999994v16.71666666666667h-3.3599999999999994z m5-13.360000000000001h3.3599999999999994v10.000000000000002h-3.3599999999999994v-10z m-31.64 13.360000000000001v-16.71666666666667h3.3599999999999994v16.71666666666667h-3.3599999999999994z m-5-3.3599999999999994v-10h3.3600000000000003v10h-3.3600000000000003z' })
                )
            );
        }
    }]);

    return MdVibration;
}(React.Component);

exports.default = MdVibration;
module.exports = exports['default'];