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

var MdPhonelink = function (_React$Component) {
    _inherits(MdPhonelink, _React$Component);

    function MdPhonelink() {
        _classCallCheck(this, MdPhonelink);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdPhonelink).apply(this, arguments));
    }

    _createClass(MdPhonelink, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm36.64000000000001 28.36v-11.716666666666669h-6.640000000000008v11.716666666666669h6.640000000000001z m1.7199999999999989-15q0.7033333333333331 0 1.1716666666666669 0.4666666666666668t0.4683333333333266 1.1733333333333338v16.64q0 0.7033333333333331-0.46666666666666856 1.211666666666666t-1.173333333333332 0.509999999999998h-10q-0.7033333333333331 0-1.211666666666666-0.5083333333333329t-0.5100000000000016-1.2100000000000009v-16.64333333333333q0-0.7033333333333331 0.5083333333333329-1.1716666666666669t1.2100000000000009-0.4666666666666668h10.000000000000004z m-31.720000000000006-3.3599999999999994v18.36h16.716666666666665v5h-23.356666666666666v-5h3.358333333333334v-18.36q0-1.3283333333333331 0.976666666666667-2.3433333333333337t2.3066666666666666-1.0166666666666666h30v3.3616666666666672h-30z' })
                )
            );
        }
    }]);

    return MdPhonelink;
}(React.Component);

exports.default = MdPhonelink;
module.exports = exports['default'];