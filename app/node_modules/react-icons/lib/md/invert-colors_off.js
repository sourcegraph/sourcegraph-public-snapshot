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

var MdInvertColorsOff = function (_React$Component) {
    _inherits(MdInvertColorsOff, _React$Component);

    function MdInvertColorsOff() {
        _classCallCheck(this, MdInvertColorsOff);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdInvertColorsOff).apply(this, arguments));
    }

    _createClass(MdInvertColorsOff, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm20 8.516666666666667l-3.828333333333333 3.75-2.343333333333332-2.3466666666666676 6.171666666666665-6.17 9.453333333333333 9.453333333333333q2.9666666666666686 2.9666666666666686 3.671666666666667 7.149999999999999t-1.0166666666666657 7.850000000000001l-12.108333333333334-12.033333333333331v-7.653333333333334z m0 24.14v-8.046666666666667l-7.966666666666669-7.966666666666669q-2.0333333333333314 2.653333333333336-2.0333333333333314 6.01166666666667 0 4.063333333333333 2.966666666666667 7.033333333333331t7.033333333333333 2.9666666666666686z m14.453333333333333 2.1083333333333343l0.5466666666666669 0.6233333333333348-2.1099999999999994 2.1099999999999994-4.533333333333335-4.533333333333331q-3.6683333333333294 2.973333333333329-8.356666666666666 2.973333333333329-5.550000000000001 0-9.455-3.9066666666666663-3.5933333333333337-3.671666666666667-3.866666666666667-8.788333333333334t2.9299999999999997-9.026666666666666l-4.6083333333333325-4.606666666666666 2.1066666666666674-2.110000000000001q23.36 23.356666666666666 27.343333333333334 27.26166666666667z' })
                )
            );
        }
    }]);

    return MdInvertColorsOff;
}(React.Component);

exports.default = MdInvertColorsOff;
module.exports = exports['default'];