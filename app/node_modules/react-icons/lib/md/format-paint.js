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

var MdFormatPaint = function (_React$Component) {
    _inherits(MdFormatPaint, _React$Component);

    function MdFormatPaint() {
        _classCallCheck(this, MdFormatPaint);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdFormatPaint).apply(this, arguments));
    }

    _createClass(MdFormatPaint, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm30 6.640000000000001h5v13.36h-13.36v15q0 0.7033333333333331-0.466666666666665 1.1716666666666669t-1.173333333333332 0.46666666666666856h-3.360000000000003q-0.7033333333333331 0-1.1716666666666669-0.46666666666666856t-0.4683333333333337-1.1716666666666669v-18.36h16.64v-6.640000000000001h-1.6400000000000006v1.6400000000000006q0 0.7033333333333331-0.466666666666665 1.211666666666666t-1.173333333333332 0.5099999999999998h-20q-0.7033333333333331 0-1.211666666666667-0.5083333333333329t-0.5099999999999998-1.209999999999999v-6.643333333333334q0-0.7033333333333331 0.5083333333333337-1.1716666666666669t1.209999999999999-0.4666666666666668h20q0.7033333333333331 0 1.1716666666666669 0.4666666666666668t0.471666666666664 1.1716666666666669v1.6400000000000006z' })
                )
            );
        }
    }]);

    return MdFormatPaint;
}(React.Component);

exports.default = MdFormatPaint;
module.exports = exports['default'];