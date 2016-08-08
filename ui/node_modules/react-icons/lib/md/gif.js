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

var MdGif = function (_React$Component) {
    _inherits(MdGif, _React$Component);

    function MdGif() {
        _classCallCheck(this, MdGif);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdGif).apply(this, arguments));
    }

    _createClass(MdGif, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm31.640000000000004 17.5h-5v1.6400000000000006h3.359999999999996v2.5h-3.3599999999999994v3.3599999999999994h-2.5v-10h7.5v2.5z m-16.640000000000004-2.5q0.7033333333333331 0 1.1716666666666669 0.4666666666666668t0.466666666666665 1.1733333333333338v0.8599999999999994h-5.778333333333331v5h3.283333333333333v-2.5h2.5v3.3599999999999994q0 0.7033333333333331-0.46999999999999886 1.1716666666666669t-1.1733333333333356 0.4683333333333337h-5q-0.7033333333333331 0-1.1716666666666669-0.466666666666665t-0.4666666666666668-1.173333333333332v-6.716666666666669q0-0.7050000000000001 0.4666666666666668-1.1733333333333338t1.1716666666666669-0.466666666666665h5z m4.140000000000001 0h2.5v10h-2.5v-10z' })
                )
            );
        }
    }]);

    return MdGif;
}(React.Component);

exports.default = MdGif;
module.exports = exports['default'];