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

var FaCameraRetro = function (_React$Component) {
    _inherits(FaCameraRetro, _React$Component);

    function FaCameraRetro() {
        _classCallCheck(this, FaCameraRetro);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaCameraRetro).apply(this, arguments));
    }

    _createClass(FaCameraRetro, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm20.714285714285715 18.571428571428573q0-0.31428571428571317-0.1999999999999993-0.514285714285716t-0.514285714285716-0.1999999999999993q-1.4714285714285715 0-2.5228571428571414 1.048571428571428t-1.048571428571428 2.522857142857145q0 0.31428571428571317 0.1999999999999993 0.5142857142857125t0.514285714285716 0.1999999999999993 0.5142857142857125-0.1999999999999993 0.1999999999999993-0.5142857142857125q0-0.8928571428571423 0.6257142857142846-1.5171428571428578t1.5171428571428578-0.6257142857142881q0.31428571428571317 0 0.5142857142857125-0.1999999999999993t0.1999999999999993-0.514285714285716z m5 2.8999999999999986q0 2.3671428571428557-1.6742857142857126 4.042857142857141t-4.040000000000003 1.6714285714285744-4.039999999999999-1.6714285714285708-1.6742857142857144-4.0428571428571445 1.6742857142857144-4.0385714285714265 4.039999999999999-1.6757142857142888 4.039999999999999 1.6757142857142853 1.6742857142857162 4.03857142857143z m-22.857142857142858 12.814285714285713h34.28571428571428v-2.8571428571428577h-34.285714285714285v2.8571428571428577z m25.714285714285715-12.814285714285717q0-3.547142857142859-2.5114285714285707-6.057142857142857t-6.060000000000002-2.5142857142857125-6.0600000000000005 2.5142857142857142-2.5114285714285707 6.057142857142859 2.5114285714285725 6.061428571428571 6.059999999999999 2.5100000000000016 6.060000000000002-2.509999999999998 2.5114285714285707-6.061428571428575z m-22.857142857142858-14.328571428571426h8.57142857142857v-2.8571428571428568h-8.57142857142857v2.8571428571428568z m-2.857142857142858 4.2857142857142865h34.285714285714285v-5.714285714285714h-18.48285714285714l-1.428571428571427 2.8571428571428568h-14.374285714285715v2.8571428571428577z m37.142857142857146-5.714285714285714v28.57142857142857q0 1.182857142857145-0.8371428571428581 2.020000000000003t-2.019999999999996 0.8371428571428581h-34.28571428571429q-1.1828571428571397 0-2.019999999999997-0.8371428571428581t-0.8371428571428572-2.020000000000003v-28.57142857142857q0-1.1828571428571433 0.8371428571428571-2.020000000000001t2.02-0.8371428571428572h34.285714285714285q1.182857142857145 0 2.020000000000003 0.8371428571428572t0.8371428571428581 2.02z' })
                )
            );
        }
    }]);

    return FaCameraRetro;
}(React.Component);

exports.default = FaCameraRetro;
module.exports = exports['default'];