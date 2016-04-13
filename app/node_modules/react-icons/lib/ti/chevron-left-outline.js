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

var TiChevronLeftOutline = function (_React$Component) {
    _inherits(TiChevronLeftOutline, _React$Component);

    function TiChevronLeftOutline() {
        _classCallCheck(this, TiChevronLeftOutline);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiChevronLeftOutline).apply(this, arguments));
    }

    _createClass(TiChevronLeftOutline, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm21.666666666666668 33.333333333333336c-1.336666666666666 0-2.591666666666665-0.5200000000000031-3.5366666666666653-1.4666666666666686l-11.866666666666669-11.866666666666667 11.866666666666669-11.866666666666667c1.8900000000000006-1.8899999999999997 5.183333333333334-1.8899999999999997 7.073333333333334 0 0.9433333333333316 0.9400000000000013 1.4633333333333312 2.200000000000001 1.4633333333333312 3.533333333333335s-0.5199999999999996 2.591666666666667-1.466666666666665 3.536666666666667l-4.7933333333333366 4.796666666666665 4.796666666666667 4.800000000000001c0.9466666666666654 0.9416666666666664 1.4666666666666686 2.1999999999999993 1.4666666666666686 3.533333333333335s-0.5199999999999996 2.5916666666666686-1.4666666666666686 3.5366666666666653c-0.9416666666666664 0.9433333333333316-2.1966666666666654 1.4633333333333347-3.533333333333335 1.4633333333333347z m-10.691666666666666-13.333333333333336l9.513333333333335 9.511666666666667c0.629999999999999 0.629999999999999 1.7283333333333317 0.6283333333333339 2.3566666666666656 0 0.31666666666666643-0.31666666666666643 0.4883333333333333-0.7333333333333343 0.4883333333333333-1.1783333333333346s-0.173333333333332-0.8633333333333333-0.4883333333333333-1.1783333333333346l-7.153333333333336-7.154999999999998 7.153333333333332-7.155000000000001c0.31666666666666643-0.31666666666666643 0.4883333333333333-0.7333333333333343 0.4883333333333333-1.1783333333333328s-0.173333333333332-0.8633333333333333-0.4883333333333333-1.1783333333333328c-0.629999999999999-0.6333333333333329-1.7283333333333317-0.6300000000000008-2.3566666666666656 0l-9.513333333333334 9.511666666666667z' })
                )
            );
        }
    }]);

    return TiChevronLeftOutline;
}(React.Component);

exports.default = TiChevronLeftOutline;
module.exports = exports['default'];