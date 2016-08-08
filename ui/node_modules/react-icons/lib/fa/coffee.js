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

var FaCoffee = function (_React$Component) {
    _inherits(FaCoffee, _React$Component);

    function FaCoffee() {
        _classCallCheck(this, FaCoffee);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaCoffee).apply(this, arguments));
    }

    _createClass(FaCoffee, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm35.862068965517246 14.482758620689657q0-1.7241379310344822-1.2068965517241352-2.931034482758621t-2.9310344827586263-1.2068965517241388h-1.3793103448275872v8.275862068965516h1.3793103448275872q1.724137931034484 0 2.931034482758619-1.2068965517241388t1.2068965517241423-2.9310344827586174z m-35.862068965517246 16.551724137931032h38.62068965517241q0 2.2841379310344863-1.6165517241379277 3.900689655172414t-3.900689655172414 1.6165517241379348h-27.586206896551726q-2.2841379310344827 0-3.900689655172414-1.6165517241379277t-1.6165517241379312-3.900689655172421z m40-16.551724137931036q0 3.426206896551726-2.424827586206895 5.851034482758621t-5.851034482758621 2.4248275862068986h-1.3793103448275872v0.6896551724137936q0 1.982068965517243-1.4206896551724135 3.406896551724138t-3.406896551724138 1.4206896551724135h-15.172413793103448q-1.9820689655172412 0-3.406896551724138-1.4206896551724135t-1.4206896551724135-3.406896551724138v-15.862068965517244q0-0.5600000000000005 0.40965517241379334-0.969655172413793t0.969655172413793-0.40965517241379246h24.827586206896555q3.4262068965517223 0 5.8510344827586245 2.424827586206897t2.424827586206888 5.851034482758621z' })
                )
            );
        }
    }]);

    return FaCoffee;
}(React.Component);

exports.default = FaCoffee;
module.exports = exports['default'];