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

var MdStraighten = function (_React$Component) {
    _inherits(MdStraighten, _React$Component);

    function MdStraighten() {
        _classCallCheck(this, MdStraighten);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdStraighten).apply(this, arguments));
    }

    _createClass(MdStraighten, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm35 26.64v-13.283333333333333h-3.3599999999999994v6.643333333333333h-3.2833333333333314v-6.643333333333334h-3.356666666666669v6.643333333333334h-3.361666666666668v-6.643333333333334h-3.283333333333335v6.643333333333334h-3.354999999999997v-6.643333333333334h-3.3633333333333333v6.643333333333334h-3.283333333333333v-6.643333333333334h-3.3566666666666665v13.283333333333335h30.000000000000004z m0-16.64q1.3283333333333331 0 2.3433333333333337 1.0166666666666657t1.0166666666666657 2.341666666666667v13.283333333333333q0 1.326666666666668-1.0166666666666657 2.3416666666666686t-2.3433333333333337 1.0166666666666657h-30q-1.3283333333333331 0-2.3433333333333333-1.0166666666666657t-1.0166666666666666-2.3433333333333337v-13.283333333333333q0-1.3266666666666662 1.0166666666666666-2.341666666666667t2.3433333333333333-1.0150000000000006h30z' })
                )
            );
        }
    }]);

    return MdStraighten;
}(React.Component);

exports.default = MdStraighten;
module.exports = exports['default'];