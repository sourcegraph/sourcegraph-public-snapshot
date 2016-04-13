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

var MdPartyMode = function (_React$Component) {
    _inherits(MdPartyMode, _React$Component);

    function MdPartyMode() {
        _classCallCheck(this, MdPartyMode);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdPartyMode).apply(this, arguments));
    }

    _createClass(MdPartyMode, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm20 28.36q3.4383333333333326 0 5.899999999999999-2.461666666666666t2.4583333333333357-5.898333333333333q0-0.5466666666666669-0.1566666666666663-1.6400000000000006h-3.5166666666666657q0.31666666666666643 1.0933333333333337 0.31666666666666643 1.6400000000000006 0 2.0333333333333314-1.4866666666666681 3.5166666666666657t-3.5150000000000006 1.4833333333333343h-6.636666666666665q2.576666666666666 3.3599999999999994 6.636666666666665 3.3599999999999994z m0-16.720000000000002q-3.4383333333333326 0-5.9 2.461666666666666t-2.458333333333332 5.898333333333337q0 0.5466666666666669 0.1566666666666663 1.6400000000000006h3.5166666666666675q-0.31666666666666643-1.0933333333333337-0.31666666666666643-1.6400000000000006 0-2.0333333333333314 1.4866666666666681-3.5166666666666657t3.514999999999997-1.4833333333333343h6.638333333333335q-2.578333333333333-3.3599999999999994-6.638333333333335-3.3599999999999994z m13.36-5q1.3283333333333331 0 2.3049999999999997 1.0166666666666666t0.9750000000000014 2.3416666666666677v20q0 1.326666666666668-0.9766666666666666 2.341666666666665t-2.306666666666665 1.0166666666666657h-26.713333333333335q-1.330000000000001 0-2.3066666666666675-1.0166666666666657t-0.9766666666666666-2.3399999999999963v-20q0-1.33 0.9766666666666666-2.3450000000000006t2.3050000000000006-1.0166666666666666h5.313333333333334l3.044999999999998-3.2783333333333324h10l3.0500000000000007 3.283333333333333h5.311666666666664z' })
                )
            );
        }
    }]);

    return MdPartyMode;
}(React.Component);

exports.default = MdPartyMode;
module.exports = exports['default'];