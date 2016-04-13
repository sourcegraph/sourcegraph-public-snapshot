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

var MdExtension = function (_React$Component) {
    _inherits(MdExtension, _React$Component);

    function MdExtension() {
        _classCallCheck(this, MdExtension);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdExtension).apply(this, arguments));
    }

    _createClass(MdExtension, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm34.14000000000001 18.36q1.7166666666666686 0 2.9666666666666686 1.211666666666666t1.25 2.9299999999999997-1.25 2.9333333333333336-2.9666666666666686 1.2100000000000009h-2.5v6.716666666666665q0 1.3299999999999983-0.9766666666666666 2.306666666666665t-2.3049999999999997 0.9766666666666666h-6.328333333333333v-2.5q0-1.875-1.3283333333333331-3.163333333333334t-3.1999999999999993-1.288333333333334-3.205 1.288333333333334-1.3283333333333331 3.163333333333334v2.5h-6.330000000000009q-1.3283333333333331 0-2.3049999999999997-0.9766666666666666t-0.9766666666666666-2.3049999999999997v-6.329999999999998h2.5q1.875 0 3.163333333333332-1.3283333333333331t1.288333333333334-3.203333333333333-1.288333333333334-3.1999999999999993-3.163333333333333-1.3300000000000018h-2.5v-6.3299999999999965q0-1.3283333333333331 0.9766666666666666-2.3049999999999997t2.3050000000000006-0.9766666666666666h6.716666666666667v-2.5q0-1.7166666666666668 1.213333333333333-2.966666666666667t2.933333333333332-1.25 2.9283333333333346 1.25 1.211666666666666 2.966666666666667v2.5h6.716666666666669q1.3299999999999983 0 2.306666666666665 0.9766666666666666t0.9766666666666666 2.3049999999999997v6.716666666666667h2.5z' })
                )
            );
        }
    }]);

    return MdExtension;
}(React.Component);

exports.default = MdExtension;
module.exports = exports['default'];