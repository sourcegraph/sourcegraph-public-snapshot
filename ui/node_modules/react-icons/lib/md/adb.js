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

var MdAdb = function (_React$Component) {
    _inherits(MdAdb, _React$Component);

    function MdAdb() {
        _classCallCheck(this, MdAdb);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdAdb).apply(this, arguments));
    }

    _createClass(MdAdb, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm25 15q0.7033333333333331 0 1.1716666666666669-0.4666666666666668t0.466666666666665-1.1733333333333338-0.466666666666665-1.211666666666666-1.1716666666666669-0.5099999999999998-1.1716666666666669 0.5083333333333329-0.466666666666665 1.209999999999999 0.466666666666665 1.1716666666666669 1.1716666666666669 0.47166666666666757z m-10 0q0.7033333333333331 0 1.1716666666666669-0.4666666666666668t0.466666666666665-1.1733333333333338-0.466666666666665-1.211666666666666-1.1716666666666669-0.5099999999999998-1.1716666666666669 0.5083333333333329-0.4666666666666668 1.209999999999999 0.4666666666666668 1.1716666666666669 1.1716666666666669 0.47166666666666757z m11.875-7.733333333333333q4.766666666666666 3.5133333333333345 4.766666666666666 9.373333333333335v1.7166666666666686h-23.284999999999997v-1.7166666666666686q-1.7763568394002505e-15-5.859999999999999 4.766666666666666-9.375l-3.5166666666666657-3.5166666666666666 1.4066666666666663-1.3283333333333331 3.8283333333333314 3.829999999999999q2.501666666666667-1.25 5.158333333333333-1.25t5.156666666666666 1.25l3.828333333333333-3.8283333333333336 1.408333333333335 1.3283333333333336z m-18.516666666666666 19.37166666666667v-6.638333333333335h23.28333333333333v6.640000000000001q0 4.843333333333334-3.3999999999999986 8.283333333333331t-8.240000000000002 3.4366666666666674-8.24-3.4383333333333326-3.4000000000000004-8.283333333333331z' })
                )
            );
        }
    }]);

    return MdAdb;
}(React.Component);

exports.default = MdAdb;
module.exports = exports['default'];