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

var MdLoyalty = function (_React$Component) {
    _inherits(MdLoyalty, _React$Component);

    function MdLoyalty() {
        _classCallCheck(this, MdLoyalty);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdLoyalty).apply(this, arguments));
    }

    _createClass(MdLoyalty, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm28.75 25.466666666666665q1.25-1.25 1.25-2.966666666666665t-1.211666666666666-2.9299999999999997-2.9299999999999997-1.211666666666666q-1.7966666666666669 0-2.9666666666666686 1.1716666666666669l-1.25 1.25-1.173333333333332-1.25q-1.1700000000000017-1.1716666666666669-2.9666666666666686-1.1716666666666669-1.7166666666666668 0-2.9299999999999997 1.211666666666666t-1.2133333333333312 2.9299999999999997q0 1.7966666666666669 1.1716666666666669 2.9666666666666686l7.109999999999999 7.111666666666665z m-19.61-13.824999999999998q1.0166666666666657 0 1.7583333333333329-0.7416666666666671t0.7400000000000002-1.7599999999999998-0.7383333333333333-1.7566666666666677-1.7599999999999998-0.7433333333333332-1.7583333333333329 0.7416666666666663-0.7433333333333332 1.756666666666666 0.7416666666666663 1.7583333333333329 1.756666666666666 0.7400000000000002z m26.563333333333333 7.656666666666668q0.9383333333333326 0.9383333333333326 0.9383333333333326 2.3433333333333337t-0.9383333333333326 2.3433333333333337l-11.716666666666669 11.716666666666665q-0.9400000000000013 0.9399999999999977-2.344999999999999 0.9399999999999977t-2.3433333333333337-0.9383333333333326l-15-15q-0.94-0.9366666666666674-0.94-2.3416666666666686v-11.716666666666667q0-1.33 0.976666666666667-2.3066666666666666t2.3049999999999997-0.9750000000000001h11.716666666666667q1.408333333333335 0 2.344999999999999 0.938333333333333z' })
                )
            );
        }
    }]);

    return MdLoyalty;
}(React.Component);

exports.default = MdLoyalty;
module.exports = exports['default'];