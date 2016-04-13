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

var MdCompare = function (_React$Component) {
    _inherits(MdCompare, _React$Component);

    function MdCompare() {
        _classCallCheck(this, MdCompare);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdCompare).apply(this, arguments));
    }

    _createClass(MdCompare, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm31.640000000000004 5q1.3283333333333367 0 2.34333333333333 1.0166666666666666t1.0166666666666657 2.3400000000000007v23.28333333333333q0 1.326666666666668-1.0166666666666657 2.3416666666666686t-2.3433333333333337 1.0166666666666657h-8.283333333333331v-15l8.283333333333331 10v-21.641666666666666h-8.283333333333331v-3.356666666666669h8.283333333333331z m-15 25v-10l-8.283333333333333 10h8.283333333333333z m0-25v-3.36h3.359999999999996v36.71666666666667h-3.3599999999999994v-3.356666666666669h-8.283333333333333q-1.326666666666667 0-2.341666666666667-1.0166666666666657t-1.0150000000000006-2.34333333333333v-23.28333333333334q0-1.3266666666666653 1.0166666666666666-2.341666666666665t2.3400000000000007-1.0150000000000006h8.283333333333333z' })
                )
            );
        }
    }]);

    return MdCompare;
}(React.Component);

exports.default = MdCompare;
module.exports = exports['default'];