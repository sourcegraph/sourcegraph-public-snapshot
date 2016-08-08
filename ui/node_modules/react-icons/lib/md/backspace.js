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

var MdBackspace = function (_React$Component) {
    _inherits(MdBackspace, _React$Component);

    function MdBackspace() {
        _classCallCheck(this, MdBackspace);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdBackspace).apply(this, arguments));
    }

    _createClass(MdBackspace, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm31.640000000000004 26.016666666666666l-5.940000000000005-6.016666666666666 5.940000000000001-6.016666666666666-2.3433333333333337-2.341666666666667-5.938333333333333 6.0166666666666675-6.016666666666666-6.016666666666666-2.3416666666666686 2.341666666666665 6.016666666666666 6.016666666666666-6.016666666666666 6.016666666666666 2.3433333333333337 2.3416666666666686 6.016666666666666-6.016666666666666 5.936666666666667 6.016666666666666z m4.9999999999999964-21.016666666666666q1.3283333333333331 0 2.3433333333333337 1.0166666666666666t1.0166666666666657 2.3400000000000007v23.28333333333333q0 1.326666666666668-1.0166666666666657 2.3416666666666686t-2.3433333333333337 1.0166666666666657h-25q-1.5633333333333326 0-2.6566666666666663-1.4866666666666646l-8.983333333333334-13.515000000000004 8.983333333333333-13.513333333333332q1.0933333333333355-1.4833333333333343 2.656666666666668-1.4833333333333343h25z' })
                )
            );
        }
    }]);

    return MdBackspace;
}(React.Component);

exports.default = MdBackspace;
module.exports = exports['default'];