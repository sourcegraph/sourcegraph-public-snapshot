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

var TiNotes = function (_React$Component) {
    _inherits(TiNotes, _React$Component);

    function TiNotes() {
        _classCallCheck(this, TiNotes);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiNotes).apply(this, arguments));
    }

    _createClass(TiNotes, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm31.383333333333333 6.766666666666667c-0.17666666666666586-0.16000000000000014-0.40333333333333243-0.2333333333333334-0.6550000000000011-0.20333333333333314l-18.333333333333336 2.083333333333333c-0.4166666666666661 0.05000000000000071-0.7283333333333335 0.4066666666666663-0.7283333333333335 0.8266666666666662v17.193333333333335c-2.756666666666666 0-5 1.870000000000001-5 4.166666666666668s2.2433333333333376 4.166666666666664 5.000000000000005 4.166666666666664 5-1.8699999999999974 5-4.166666666666668v-12.683333333333334l10-1.0399999999999991v6.223333333333336c-2.7566666666666677 0-5 1.870000000000001-5 4.166666666666668s2.2433333333333323 4.166666666666668 5 4.166666666666668 5-1.870000000000001 5-4.166666666666668v-20.110000000000003c0-0.2400000000000002-0.10333333333333172-0.4666666666666668-0.283333333333335-0.625z' })
                )
            );
        }
    }]);

    return TiNotes;
}(React.Component);

exports.default = TiNotes;
module.exports = exports['default'];