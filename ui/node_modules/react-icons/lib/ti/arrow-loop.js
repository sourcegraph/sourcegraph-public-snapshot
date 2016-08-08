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

var TiArrowLoop = function (_React$Component) {
    _inherits(TiArrowLoop, _React$Component);

    function TiArrowLoop() {
        _classCallCheck(this, TiArrowLoop);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiArrowLoop).apply(this, arguments));
    }

    _createClass(TiArrowLoop, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm27.5 13.333333333333334h-3.4766666666666666l2.155000000000001-2.1549999999999994c0.6499999999999986-0.6500000000000004 0.6499999999999986-1.705 0-2.3566666666666656s-1.7049999999999983-0.6500000000000004-2.3566666666666656 0l-6.178333333333335 6.178333333333331 6.178333333333335 6.178333333333335c0.3249999999999993 0.3249999999999993 0.75 0.4883333333333333 1.1783333333333346 0.4883333333333333s0.8533333333333317-0.163333333333334 1.1783333333333346-0.4883333333333333c0.6499999999999986-0.6499999999999986 0.6499999999999986-1.7049999999999983 0-2.3566666666666656l-2.1550000000000047-2.155000000000001h3.4766666666666666c2.3000000000000007 0 4.166666666666668 2.2433333333333323 4.166666666666668 5s-2.2433333333333323 5-5 5h-13.333333333333334c-2.756666666666666 0-5-2.2433333333333323-5-5s2.243333333333334-5 5-5c0.9216666666666669 0 1.666666666666666-0.7466666666666661 1.666666666666666-1.666666666666666s-0.7449999999999992-1.666666666666666-1.666666666666666-1.666666666666666c-4.595000000000001 0-8.333333333333334 3.7383333333333333-8.333333333333334 8.333333333333336s3.7383333333333333 8.333333333333336 8.333333333333334 8.333333333333336h13.333333333333334c4.595000000000002 0 8.333333333333332-3.7383333333333333 8.333333333333332-8.333333333333336s-3.366666666666667-8.333333333333334-7.5-8.333333333333334z' })
                )
            );
        }
    }]);

    return TiArrowLoop;
}(React.Component);

exports.default = TiArrowLoop;
module.exports = exports['default'];