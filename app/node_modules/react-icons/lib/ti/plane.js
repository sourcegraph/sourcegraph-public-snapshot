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

var TiPlane = function (_React$Component) {
    _inherits(TiPlane, _React$Component);

    function TiPlane() {
        _classCallCheck(this, TiPlane);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiPlane).apply(this, arguments));
    }

    _createClass(TiPlane, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm33.32666666666667 22.511666666666667l-9.993333333333332-5.711666666666666v-9.923333333333332c0-1.3783333333333339-1.1216666666666661-2.500000000000001-2.5-2.500000000000001s-2.5 1.121666666666667-2.5 2.5v9.926666666666666l-9.993333333333334 5.711666666666666c-0.7333333333333334 0.41666666666666785-1.036666666666667 1.3166666666666664-0.71 2.0933333333333337s1.1833333333333327 1.1883333333333326 1.9933333333333332 0.9549999999999983l8.71-2.4899999999999984v7.416666666666668l-2.708333333333334 2.1666666666666643c-0.6449999999999996 0.5166666666666657-0.8133333333333326 1.4283333333333346-0.40000000000000036 2.1416666666666657s1.2950000000000017 1.0116666666666703 2.0599999999999987 0.7066666666666634l3.5500000000000007-1.4200000000000017 3.5500000000000007 1.4200000000000017c0.1999999999999993 0.0799999999999983 0.41000000000000014 0.11666666666666714 0.6166666666666671 0.11666666666666714 0.5783333333333331 0 1.1333333333333329-0.29999999999999716 1.4400000000000013-0.826666666666668 0.41666666666666785-0.7133333333333312 0.245000000000001-1.625-0.3999999999999986-2.1400000000000006l-2.7083333333333357-2.164999999999992v-7.416666666666668l8.71 2.486666666666668c0.14999999999999858 0.043333333333333 0.30666666666666487 0.06666666666666643 0.45666666666666345 0.06666666666666643 0.6566666666666663 0 1.2700000000000031-0.39000000000000057 1.5366666666666688-1.0216666666666683 0.326666666666668-0.7766666666666673 0.023333333333333428-1.6750000000000007-0.7100000000000009-2.0933333333333337z m-12.493333333333332-15.219999999999999c-0.46000000000000085-8.881784197001252e-16-0.8333333333333321-0.37333333333333396-0.8333333333333321-0.8333333333333339s0.37333333333333485-0.833333333333333 0.8333333333333321-0.833333333333333 0.8333333333333321 0.3733333333333331 0.8333333333333321 0.833333333333333-0.37333333333333485 0.833333333333333-0.8333333333333321 0.833333333333333z' })
                )
            );
        }
    }]);

    return TiPlane;
}(React.Component);

exports.default = TiPlane;
module.exports = exports['default'];