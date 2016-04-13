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

var FaFighterJet = function (_React$Component) {
    _inherits(FaFighterJet, _React$Component);

    function FaFighterJet() {
        _classCallCheck(this, FaFighterJet);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaFighterJet).apply(this, arguments));
    }

    _createClass(FaFighterJet, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm40 21.333333333333332q-0.021333333333330984 0.6666666666666679-6 2l-7.333333333333336 0.6666666666666679-4.666666666666664 1.3333333333333321h-1.3333333333333321l-6.103999999999999 7.333333333333332h1.4373333333333314q0.5413333333333341 0 0.9373333333333349 0.09333333333333371t0.3960000000000008 0.240000000000002-0.3960000000000008 0.240000000000002-0.9373333333333349 0.0933333333333266h-6.666666666666668v-0.6666666666666643h1.333333333333334v-8.666666666666664h-3.333333333333333l-4 4.666666666666664h-1.9999999999999998l-0.6666666666666666-0.6666666666666643v-4h0.6666666666666666v-0.6666666666666679h2.666666666666667v-0.16666666666666785l-4-0.5v-2.666666666666668l4-0.5v-0.1666666666666643h-2.666666666666667v-0.6666666666666679h-0.6666666666666664v-4l0.6666666666666666-0.6666666666666661h1.9999999999999998l4 4.666666666666666h3.333333333333333v-8.666666666666664h-1.333333333333334v-0.6666666666666679h6.666666666666666q0.5413333333333323 0 0.9373333333333331 0.09333333333333371t0.3960000000000008 0.2400000000000002-0.3960000000000008 0.2400000000000002-0.9373333333333314 0.09333333333333371h-1.4373333333333331l6.103999999999997 7.333333333333332h1.3333333333333321l4.666666666666668 1.3333333333333321 7.333333333333336 0.6666666666666679q5.437333333333335 1.2079999999999984 5.978666666666669 1.937333333333335z' })
                )
            );
        }
    }]);

    return FaFighterJet;
}(React.Component);

exports.default = FaFighterJet;
module.exports = exports['default'];