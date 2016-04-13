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

var MdGridOn = function (_React$Component) {
    _inherits(MdGridOn, _React$Component);

    function MdGridOn() {
        _classCallCheck(this, MdGridOn);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdGridOn).apply(this, arguments));
    }

    _createClass(MdGridOn, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm33.36 13.360000000000001v-6.716666666666668h-6.716666666666669v6.716666666666668h6.716666666666669z m0 9.999999999999998v-6.716666666666669h-6.716666666666669v6.716666666666669h6.716666666666669z m0 10v-6.716666666666669h-6.716666666666669v6.716666666666669h6.716666666666669z m-10-20v-6.716666666666668h-6.716666666666669v6.716666666666668h6.716666666666669z m0 10v-6.716666666666669h-6.716666666666669v6.716666666666669h6.716666666666669z m0 10v-6.716666666666669h-6.716666666666669v6.716666666666669h6.716666666666669z m-10-20v-6.716666666666668h-6.716666666666668v6.716666666666668h6.716666666666668z m0 10v-6.716666666666669h-6.716666666666668v6.716666666666669h6.716666666666668z m0 10v-6.716666666666669h-6.716666666666668v6.716666666666669h6.716666666666668z m20-30q1.3283333333333331 0 2.3049999999999997 0.9766666666666666t0.9750000000000014 2.3049999999999997v26.71666666666667q0 1.3299999999999983-0.9766666666666666 2.306666666666665t-2.306666666666665 0.9750000000000014h-26.713333333333335q-1.330000000000001 0-2.3066666666666675-0.9766666666666666t-0.9766666666666666-2.306666666666665v-26.713333333333335q0-1.330000000000001 0.9766666666666666-2.3066666666666675t2.3050000000000006-0.9766666666666666h26.71666666666667z' })
                )
            );
        }
    }]);

    return MdGridOn;
}(React.Component);

exports.default = MdGridOn;
module.exports = exports['default'];