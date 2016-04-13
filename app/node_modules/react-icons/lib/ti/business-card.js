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

var TiBusinessCard = function (_React$Component) {
    _inherits(TiBusinessCard, _React$Component);

    function TiBusinessCard() {
        _classCallCheck(this, TiBusinessCard);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiBusinessCard).apply(this, arguments));
    }

    _createClass(TiBusinessCard, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm33.333333333333336 33.333333333333336h-26.666666666666668c-2.7566666666666677 0-5.000000000000001-2.2433333333333323-5.000000000000001-5v-16.666666666666668c0-2.7566666666666677 2.2433333333333336-5 5-5h26.666666666666668c2.7566666666666677 0 5 2.243333333333334 5 5v16.666666666666668c0 2.7566666666666677-2.2433333333333323 5-5 5z m-26.666666666666668-23.333333333333336c-0.9166666666666679 0-1.6666666666666679 0.75-1.6666666666666679 1.666666666666666v16.66666666666667c0 0.9166666666666679 0.75 1.6666666666666679 1.666666666666667 1.6666666666666679h26.666666666666668c0.9166666666666643 0 1.6666666666666643-0.75 1.6666666666666643-1.6666666666666679v-16.666666666666668c0-0.9166666666666661-0.75-1.666666666666666-1.6666666666666643-1.666666666666666h-26.666666666666668z m10 15h-6.666666666666668c-0.9216666666666669 0-1.666666666666666-0.7466666666666661-1.666666666666666-1.6666666666666679s0.7449999999999992-1.6666666666666679 1.666666666666666-1.6666666666666679h6.666666666666668c0.9216666666666669 0 1.6666666666666679 0.7466666666666661 1.6666666666666679 1.6666666666666679s-0.745000000000001 1.6666666666666679-1.6666666666666679 1.6666666666666679z m0-6.666666666666664h-6.666666666666668c-0.9216666666666669 0-1.666666666666666-0.7466666666666661-1.666666666666666-1.6666666666666679s0.7449999999999992-1.666666666666666 1.666666666666666-1.666666666666666h6.666666666666668c0.9216666666666669 0 1.6666666666666679 0.7466666666666661 1.6666666666666679 1.666666666666666s-0.745000000000001 1.6666666666666679-1.6666666666666679 1.6666666666666679z m13.333333333333332-0.8333333333333357c0 1.8416666666666686-1.4916666666666671 3.333333333333332-3.333333333333332 3.333333333333332s-3.333333333333332-1.4916666666666671-3.333333333333332-3.333333333333332 1.4916666666666671-3.333333333333334 3.333333333333332-3.333333333333334 3.333333333333332 1.4916666666666671 3.333333333333332 3.333333333333334z m-3.333333333333332 4.760000000000002c-2.6033333333333353 0-4.166666666666668 1.1916666666666664-4.166666666666668 2.383333333333333 0 0.5933333333333337 1.5633333333333326 1.1900000000000013 4.166666666666668 1.1900000000000013 2.4433333333333316 0 4.166666666666668-0.5949999999999989 4.166666666666668-1.1916666666666664 0-1.1900000000000013-1.6333333333333329-2.383333333333333-4.166666666666668-2.383333333333333z' })
                )
            );
        }
    }]);

    return TiBusinessCard;
}(React.Component);

exports.default = TiBusinessCard;
module.exports = exports['default'];