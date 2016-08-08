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

var TiWorld = function (_React$Component) {
    _inherits(TiWorld, _React$Component);

    function TiWorld() {
        _classCallCheck(this, TiWorld);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiWorld).apply(this, arguments));
    }

    _createClass(TiWorld, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm20 3.3333333333333335c-8.283333333333333 0-15 6.716666666666667-15 14.999999999999998s6.716666666666669 14.999999999999996 15 14.999999999999996 15-6.716666666666669 15-15-6.716666666666669-15-15-15z m3.333333333333332 3.3333333333333335c0 1.666666666666667-0.8333333333333321 3.333333333333333-2.5 3.333333333333333s-2.4999999999999964 1.6666666666666679-2.4999999999999964 3.333333333333334v5.000000000000002s1.6666666666666679 0 1.6666666666666679-5c0-0.9216666666666669 0.745000000000001-1.666666666666666 1.6666666666666679-1.666666666666666s1.6666666666666679 0.7449999999999992 1.6666666666666679 1.666666666666666v5c-0.9200000000000017 0-1.6666666666666679 0.7466666666666661-1.6666666666666679 1.6666666666666679s0.7466666666666661 1.6666666666666679 1.6666666666666679 1.6666666666666679c0.9216666666666669 0 1.6666666666666679-0.7466666666666661 1.6666666666666679-1.6666666666666679h1.6666666666666679v-3.333333333333332l1.6666666666666679 1.6666666666666679-1.6666666666666679 1.6666666666666679c0 5 0 5-3.333333333333332 6.666666666666668 0-1.6666666666666679-1.6666666666666679-1.6666666666666679-5-1.6666666666666679v-3.333333333333332l-3.333333333333334-3.333333333333332v-3.333333333333343c-1.666666666666666 0-1.666666666666666 1.6666666666666679-1.666666666666666 1.6666666666666679l-4.916666666666677-4.916666666666668c0.18333333333333357-0.3216666666666672 0.3733333333333331-0.6383333333333336 0.5833333333333339-0.9416666666666664l0.870000000000001-1.1300000000000008c2.4466666666666654-2.8616666666666664 6.073333333333334-4.678333333333333 10.129999999999999-4.678333333333333 1.1499999999999986 0 2.2666666666666657 0.1633333333333331 3.333333333333332 0.43666666666666654v1.2300000000000004z' })
                )
            );
        }
    }]);

    return TiWorld;
}(React.Component);

exports.default = TiWorld;
module.exports = exports['default'];