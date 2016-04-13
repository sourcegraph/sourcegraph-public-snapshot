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

var MdDirectionsTransit = function (_React$Component) {
    _inherits(MdDirectionsTransit, _React$Component);

    function MdDirectionsTransit() {
        _classCallCheck(this, MdDirectionsTransit);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdDirectionsTransit).apply(this, arguments));
    }

    _createClass(MdDirectionsTransit, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm30 18.36v-8.36h-8.36v8.36h8.36z m-2.5 10q1.0933333333333337 0 1.7966666666666669-0.7416666666666671t0.7033333333333331-1.7566666666666642-0.7033333333333331-1.7583333333333329-1.7966666666666669-0.7399999999999984-1.7966666666666669 0.7416666666666671-0.7033333333333331 1.7566666666666642 0.7033333333333331 1.7583333333333329 1.7966666666666669 0.7433333333333323z m-9.14-10v-8.36h-8.36v8.36h8.36z m-5.859999999999999 10q1.0933333333333337 0 1.7966666666666669-0.7416666666666671t0.7033333333333331-1.7566666666666642-0.7033333333333331-1.7583333333333329-1.7966666666666669-0.7399999999999984-1.7966666666666669 0.7416666666666671-0.7033333333333331 1.7566666666666642 0.7033333333333331 1.7583333333333329 1.7966666666666669 0.7433333333333323z m7.5-25q6.483333333333334 0 9.921666666666667 1.3283333333333331t3.4383333333333326 5.311666666666667v15.861666666666665q0 2.421666666666667-1.7166666666666686 4.100000000000001t-4.141666666666666 1.6833333333333336l2.5 2.4999999999999964v0.855000000000004h-20.001666666666665v-0.8599999999999994l2.5-2.5q-2.42 0-4.138333333333334-1.6799999999999997t-1.7166666666666668-4.100000000000001v-15.86q0-3.9833333333333343 3.4366666666666674-5.3116666666666665t9.918333333333333-1.3283333333333331z' })
                )
            );
        }
    }]);

    return MdDirectionsTransit;
}(React.Component);

exports.default = MdDirectionsTransit;
module.exports = exports['default'];