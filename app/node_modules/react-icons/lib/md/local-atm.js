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

var MdLocalAtm = function (_React$Component) {
    _inherits(MdLocalAtm, _React$Component);

    function MdLocalAtm() {
        _classCallCheck(this, MdLocalAtm);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdLocalAtm).apply(this, arguments));
    }

    _createClass(MdLocalAtm, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm33.36 30v-20h-26.716666666666665v20h26.716666666666665z m0-23.36q1.4066666666666663 0 2.3433333333333337 0.9766666666666666t0.9383333333333326 2.383333333333333v20q0 1.4066666666666663-0.9383333333333326 2.383333333333333t-2.3433333333333337 0.9766666666666666h-26.716666666666665q-1.408333333333334 0-2.3450000000000006-0.9766666666666666t-0.94-2.383333333333333v-20q0-1.4066666666666663 0.938333333333333-2.383333333333333t2.341666666666667-0.9766666666666666h26.71666666666667z m-15 21.720000000000002v-1.716666666666665h-3.3599999999999994v-3.2833333333333314h6.640000000000001v-1.7166666666666686h-5q-0.7033333333333331 0-1.1716666666666669-0.46999999999999886t-0.4683333333333337-1.1733333333333391v-5q0-0.6999999999999993 0.4666666666666668-1.17t1.1733333333333338-0.4666666666666668h1.7166666666666686v-1.7216666666666658h3.2833333333333314v1.7166666666666668h3.3599999999999994v3.2833333333333314h-6.640000000000001v1.7166666666666686h5q0.7033333333333331 0 1.1716666666666669 0.46999999999999886t0.4683333333333337 1.1716666666666669v5q0 0.7049999999999983-0.466666666666665 1.173333333333332t-1.173333333333332 0.466666666666665h-1.716666666666665v1.7199999999999989h-3.2833333333333314z' })
                )
            );
        }
    }]);

    return MdLocalAtm;
}(React.Component);

exports.default = MdLocalAtm;
module.exports = exports['default'];