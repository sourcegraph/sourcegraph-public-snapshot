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

var MdMovieCreation = function (_React$Component) {
    _inherits(MdMovieCreation, _React$Component);

    function MdMovieCreation() {
        _classCallCheck(this, MdMovieCreation);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdMovieCreation).apply(this, arguments));
    }

    _createClass(MdMovieCreation, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm30 6.640000000000001h6.640000000000001v23.36q0 1.3283333333333331-0.9766666666666666 2.3433333333333337t-2.3049999999999997 1.0166666666666657h-26.71666666666667q-1.3299999999999992 0-2.3066666666666658-1.0166666666666657t-0.9749999999999996-2.3433333333333337v-20q0-1.3283333333333331 0.9766666666666666-2.3433333333333337t2.3066666666666666-1.0166666666666666h1.716666666666666l3.283333333333333 6.72h4.999999999999998l-3.283333333333333-6.716666666666668h3.283333333333333l3.356666666666669 6.7150000000000025h5l-3.3599999999999994-6.716666666666668h3.3599999999999994l3.3599999999999994 6.716666666666668h5z' })
                )
            );
        }
    }]);

    return MdMovieCreation;
}(React.Component);

exports.default = MdMovieCreation;
module.exports = exports['default'];