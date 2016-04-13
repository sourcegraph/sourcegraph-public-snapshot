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

var TiTimes = function (_React$Component) {
    _inherits(TiTimes, _React$Component);

    function TiTimes() {
        _classCallCheck(this, TiTimes);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiTimes).apply(this, arguments));
    }

    _createClass(TiTimes, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm29.023333333333337 10.976666666666668c-1.3000000000000007-1.3000000000000007-3.413333333333334-1.3000000000000007-4.713333333333335 0l-4.310000000000002 4.3100000000000005-4.3100000000000005-4.3100000000000005c-1.3000000000000007-1.3000000000000007-3.413333333333334-1.3000000000000007-4.713333333333333 0-1.3000000000000007 1.3000000000000007-1.3000000000000007 3.411666666666667 0 4.713333333333333l4.306666666666667 4.309999999999999-4.306666666666667 4.309999999999999c-1.3000000000000007 1.3000000000000007-1.3000000000000007 3.4116666666666653 0 4.713333333333335 0.6500000000000004 0.6499999999999986 1.5033333333333339 0.9766666666666666 2.3566666666666656 0.9766666666666666s1.706666666666667-0.3249999999999993 2.3566666666666656-0.9766666666666666l4.310000000000002-4.309999999999999 4.309999999999999 4.309999999999999c0.6499999999999986 0.6499999999999986 1.5033333333333339 0.9766666666666666 2.3566666666666656 0.9766666666666666s1.706666666666667-0.3249999999999993 2.3566666666666656-0.9766666666666666c1.3000000000000007-1.3000000000000007 1.3000000000000007-3.411666666666669 0-4.713333333333331l-4.306666666666661-4.310000000000002 4.306666666666665-4.309999999999999c1.3000000000000007-1.3000000000000007 1.3000000000000007-3.411666666666667 0-4.713333333333333z' })
                )
            );
        }
    }]);

    return TiTimes;
}(React.Component);

exports.default = TiTimes;
module.exports = exports['default'];