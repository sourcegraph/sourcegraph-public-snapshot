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

var MdLocalOffer = function (_React$Component) {
    _inherits(MdLocalOffer, _React$Component);

    function MdLocalOffer() {
        _classCallCheck(this, MdLocalOffer);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdLocalOffer).apply(this, arguments));
    }

    _createClass(MdLocalOffer, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm9.14 11.64q1.0166666666666657 0 1.7583333333333329-0.7416666666666671t0.7400000000000002-1.7599999999999998-0.7383333333333333-1.7550000000000008-1.7599999999999998-0.7433333333333332-1.7583333333333329 0.7416666666666663-0.7433333333333332 1.756666666666666 0.7416666666666663 1.7583333333333329 1.756666666666666 0.7400000000000002z m26.563333333333333 7.656666666666666q0.9383333333333326 0.9383333333333326 0.9383333333333326 2.3433333333333337t-0.9383333333333326 2.3433333333333337l-11.716666666666669 11.716666666666669q-0.9400000000000013 0.9399999999999977-2.344999999999999 0.9399999999999977t-2.3433333333333337-0.9383333333333326l-15-15q-0.94-0.9366666666666674-0.94-2.3416666666666686v-11.716666666666667q0-1.33 0.976666666666667-2.3066666666666666t2.3049999999999997-0.9750000000000001h11.716666666666667q1.408333333333335 0 2.344999999999999 0.938333333333333z' })
                )
            );
        }
    }]);

    return MdLocalOffer;
}(React.Component);

exports.default = MdLocalOffer;
module.exports = exports['default'];