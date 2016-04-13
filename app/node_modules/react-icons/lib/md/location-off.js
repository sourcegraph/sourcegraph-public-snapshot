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

var MdLocationOff = function (_React$Component) {
    _inherits(MdLocationOff, _React$Component);

    function MdLocationOff() {
        _classCallCheck(this, MdLocationOff);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdLocationOff).apply(this, arguments));
    }

    _createClass(MdLocationOff, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm19.533333333333335 19.14q8.513333333333332 8.438333333333333 13.826666666666664 13.75l-2.1099999999999994 2.1099999999999994-5.625-5.546666666666667q-1.25 1.875-2.6566666666666663 3.671666666666667t-2.1849999999999987 2.6583333333333314l-0.783333333333335 0.8583333333333343q-0.466666666666665-0.5466666666666669-1.25-1.4450000000000003t-2.8116666666666674-3.5933333333333337-3.5533333333333346-5.233333333333334-2.7733333333333334-5.743333333333332-1.2549999999999972-5.626666666666665q0-0.8583333333333343 0.3133333333333326-2.576666666666666l-5.3133333333333335-5.3133333333333335 2.1100000000000003-2.1100000000000003 13.906666666666668 13.906666666666666z m0.466666666666665-8.28q-1.7966666666666669 0-3.046666666666667 1.4066666666666663l-5.390000000000001-5.313333333333334q3.4366666666666674-3.593333333333333 8.436666666666667-3.593333333333333 4.844999999999999 0 8.243333333333332 3.4000000000000004t3.400000000000002 8.239999999999998q0 3.75-2.8166666666666664 9.14l-6.013333333333335-6.093333333333334q1.3283333333333331-1.1716666666666669 1.3283333333333331-3.046666666666667 0-1.7166666666666668-1.211666666666666-2.9299999999999997t-2.9299999999999997-1.2116666666666678z' })
                )
            );
        }
    }]);

    return MdLocationOff;
}(React.Component);

exports.default = MdLocationOff;
module.exports = exports['default'];