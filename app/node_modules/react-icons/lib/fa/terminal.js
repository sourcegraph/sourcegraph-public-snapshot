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

var FaTerminal = function (_React$Component) {
    _inherits(FaTerminal, _React$Component);

    function FaTerminal() {
        _classCallCheck(this, FaTerminal);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaTerminal).apply(this, arguments));
    }

    _createClass(FaTerminal, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm14.485714285714288 21.942857142857143l-10.4 10.399999999999999q-0.22285714285714286 0.2228571428571442-0.5142857142857142 0.2228571428571442t-0.5114285714285716-0.2228571428571442l-1.1142857142857143-1.1142857142857139q-0.2242857142857142-0.2242857142857133-0.2242857142857142-0.5142857142857125t0.22285714285714286-0.5142857142857125l8.774285714285714-8.771428571428572-8.772857142857143-8.771428571428574q-0.22571428571428664-0.2257142857142842-0.22571428571428664-0.5142857142857142t0.22285714285714286-0.5142857142857142l1.1171428571428572-1.1142857142857139q0.22285714285714286-0.22285714285714242 0.5142857142857142-0.22285714285714242t0.5114285714285711 0.22285714285714242l10.4 10.4q0.2242857142857151 0.2242857142857133 0.2242857142857151 0.514285714285716t-0.22285714285714242 0.5142857142857125z m24.085714285714282 10.200000000000003v1.4285714285714306q0 0.3142857142857167-0.20000000000000284 0.5142857142857125t-0.5142857142857125 0.20000000000000284h-21.42857142857143q-0.31428571428571317 0-0.5142857142857142-0.20000000000000284t-0.19999999999999396-0.5142857142857196v-1.4285714285714306q0-0.31428571428571317 0.1999999999999993-0.5142857142857125t0.514285714285716-0.1999999999999993h21.42857142857143q0.3142857142857167 0 0.5142857142857125 0.1999999999999993t0.20000000000000284 0.5142857142857125z' })
                )
            );
        }
    }]);

    return FaTerminal;
}(React.Component);

exports.default = FaTerminal;
module.exports = exports['default'];