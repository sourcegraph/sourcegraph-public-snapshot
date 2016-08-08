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

var MdBuild = function (_React$Component) {
    _inherits(MdBuild, _React$Component);

    function MdBuild() {
        _classCallCheck(this, MdBuild);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdBuild).apply(this, arguments));
    }

    _createClass(MdBuild, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm37.81333333333333 31.640000000000004q0.5466666666666669 0.3133333333333326 0.5083333333333329 1.0550000000000033t-0.663333333333334 1.288333333333334l-3.828333333333333 3.828333333333333q-1.1716666666666669 1.1716666666666669-2.3433333333333337 0l-15.156666666666666-15.156666666666666q-2.8133333333333326 1.1716666666666669-5.976666666666667 0.5083333333333329t-5.508333333333334-3.008333333333333q-2.5-2.5-3.125-5.938333333333333t0.9383333333333335-6.406666666666666l7.341666666666669 7.189999999999991 5-5-7.183333333333333-7.191666666666666q2.9666666666666677-1.4066666666666667 6.406666666666665-0.8600000000000001t5.938333333333333 3.046666666666667q2.3383333333333347 2.3449999999999998 3.0050000000000026 5.510000000000001t-0.5083333333333329 5.976666666666668z' })
                )
            );
        }
    }]);

    return MdBuild;
}(React.Component);

exports.default = MdBuild;
module.exports = exports['default'];