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

var MdAttachment = function (_React$Component) {
    _inherits(MdAttachment, _React$Component);

    function MdAttachment() {
        _classCallCheck(this, MdAttachment);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdAttachment).apply(this, arguments));
    }

    _createClass(MdAttachment, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm3.3600000000000003 20.86q0-3.828333333333333 2.656666666666667-6.523333333333333t6.4833333333333325-2.6949999999999985h17.5q2.7366666666666646 0 4.689999999999998 1.9916666666666671t1.9533333333333331 4.726666666666665-1.9533333333333331 4.688333333333333-4.689999999999998 1.951666666666668h-14.136666666666665q-1.7166666666666668 0-2.966666666666667-1.2100000000000009t-1.25-2.9299999999999997 1.25-2.9666666666666686 2.966666666666667-1.25h12.500000000000002v3.356666666666669h-12.658333333333337q-0.7050000000000001 0-0.7050000000000001 0.8233333333333341t0.7050000000000001 0.8200000000000003h14.295q1.3299999999999983 0 2.344999999999999-0.9766666666666666t1.0166666666666657-2.3066666666666684-1.0166666666666657-2.3433333333333337-2.344999999999999-1.0166666666666657h-17.5q-2.416666666666666 0-4.136666666666667 1.7216666666666676t-1.7166666666666668 4.140000000000001 1.7166666666666668 4.100000000000001 4.140000000000001 1.6833333333333336h15.861666666666668v3.354999999999997h-15.865000000000002q-3.826666666666666 0-6.483333333333333-2.655000000000001t-2.6566666666666667-6.483333333333334z' })
                )
            );
        }
    }]);

    return MdAttachment;
}(React.Component);

exports.default = MdAttachment;
module.exports = exports['default'];