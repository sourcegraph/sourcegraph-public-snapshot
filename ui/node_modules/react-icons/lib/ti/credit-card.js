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

var TiCreditCard = function (_React$Component) {
    _inherits(TiCreditCard, _React$Component);

    function TiCreditCard() {
        _classCallCheck(this, TiCreditCard);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiCreditCard).apply(this, arguments));
    }

    _createClass(TiCreditCard, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm28.333333333333336 11.666666666666668h-18.333333333333336c-2.756666666666667 0-5 2.243333333333334-5 5v11.666666666666668c0 2.7566666666666677 2.243333333333334 5 5 5h18.333333333333336c2.7566666666666677 0 5-2.2433333333333323 5-5v-11.666666666666668c0-2.7566666666666677-2.2433333333333323-5-5-5z m1.6666666666666679 16.666666666666668c0 0.9200000000000017-0.7466666666666661 1.6666666666666679-1.6666666666666679 1.6666666666666679h-18.333333333333336c-0.9199999999999999 0-1.666666666666666-0.7466666666666661-1.666666666666666-1.6666666666666679v-6.666666666666668h21.666666666666664v6.666666666666668z m0-10h-21.66666666666667v-1.6666666666666679c1.7763568394002505e-15-0.9199999999999999 0.7466666666666679-1.666666666666666 1.6666666666666679-1.666666666666666h18.333333333333336c0.9200000000000017 0 1.6666666666666679 0.7466666666666661 1.6666666666666679 1.666666666666666v1.6666666666666679z m-6.666666666666668 8.333333333333332h3.333333333333332c0.46000000000000085 0 0.8333333333333321-0.37333333333333485 0.8333333333333321-0.8333333333333321s-0.37333333333333485-0.8333333333333321-0.8333333333333321-0.8333333333333321h-3.333333333333332c-0.46000000000000085 0-0.8333333333333321 0.37333333333333485-0.8333333333333321 0.8333333333333321s0.37333333333333485 0.8333333333333321 0.8333333333333321 0.8333333333333321z' })
                )
            );
        }
    }]);

    return TiCreditCard;
}(React.Component);

exports.default = TiCreditCard;
module.exports = exports['default'];