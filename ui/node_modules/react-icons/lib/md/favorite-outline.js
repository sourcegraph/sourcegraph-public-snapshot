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

var MdFavoriteOutline = function (_React$Component) {
    _inherits(MdFavoriteOutline, _React$Component);

    function MdFavoriteOutline() {
        _classCallCheck(this, MdFavoriteOutline);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdFavoriteOutline).apply(this, arguments));
    }

    _createClass(MdFavoriteOutline, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm20.156666666666666 30.938333333333333q3.75-3.3599999999999994 5.546666666666667-5.078333333333333t3.9066666666666663-4.063333333333333 2.9299999999999997-4.140000000000001 0.8200000000000003-3.5166666666666657q0-2.5-1.6799999999999997-4.138333333333334t-4.18-1.6383333333333336q-1.9533333333333331 0-3.633333333333333 1.0933333333333337t-2.3049999999999997 2.8100000000000005h-3.125q-0.625-1.7166666666666668-2.3049999999999997-2.8116666666666674t-3.6316666666666677-1.093333333333332q-2.5 0-4.183333333333332 1.6383333333333319t-1.6766666666666676 4.143333333333334q0 1.7166666666666668 0.8200000000000003 3.5166666666666657t2.9299999999999997 4.138333333333332 3.9066666666666663 4.063333333333333 5.546666666666667 5.078333333333333l0.1566666666666663 0.1566666666666663z m7.343333333333334-25.938333333333333q3.9066666666666663 0 6.523333333333333 2.6566666666666663t2.616666666666667 6.483333333333334q0 2.2666666666666657-0.8599999999999994 4.416666666666668t-3.163333333333334 4.803333333333335-4.18 4.453333333333333-6.016666666666666 5.546666666666663l-2.4200000000000017 2.1899999999999977-2.421666666666667-2.1116666666666646q-5.390000000000001-4.843333333333334-7.773333333333333-7.266666666666666t-4.413333333333334-5.699999999999999-2.033333333333333-6.33q0-3.828333333333333 2.616666666666667-6.483333333333333t6.525-2.658333333333334q4.533333333333335 0 7.5 3.5166666666666657 2.966666666666665-3.5166666666666657 7.5-3.5166666666666657z' })
                )
            );
        }
    }]);

    return MdFavoriteOutline;
}(React.Component);

exports.default = MdFavoriteOutline;
module.exports = exports['default'];