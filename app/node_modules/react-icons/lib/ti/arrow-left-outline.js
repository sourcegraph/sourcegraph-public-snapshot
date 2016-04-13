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

var TiArrowLeftOutline = function (_React$Component) {
    _inherits(TiArrowLeftOutline, _React$Component);

    function TiArrowLeftOutline() {
        _classCallCheck(this, TiArrowLeftOutline);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiArrowLeftOutline).apply(this, arguments));
    }

    _createClass(TiArrowLeftOutline, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm18.213333333333335 35c-1.3333333333333321 0-2.591666666666667-0.5200000000000031-3.533333333333333-1.4666666666666686l-11.871666666666668-11.866666666666664 11.866666666666667-11.866666666666667c1.8900000000000006-1.8916666666666666 5.183333333333334-1.8916666666666666 7.071666666666669 0 0.9416666666666664 0.9383333333333326 1.4633333333333347 2.1933333333333334 1.4633333333333347 3.5299999999999994 0 1.243333333333334-0.4499999999999993 2.416666666666666-1.2733333333333334 3.336666666666668h8.063333333333329c2.7566666666666677 0 5 2.2433333333333323 5 5s-2.2433333333333323 5-5 5h-8.060000000000002c0.8216666666666654 0.9166666666666679 1.2733333333333334 2.086666666666666 1.2733333333333334 3.330000000000002 0.0033333333333338544 1.336666666666666-0.5199999999999996 2.594999999999999-1.4666666666666686 3.539999999999999-0.9433333333333316 0.9433333333333351-2.1999999999999993 1.4633333333333312-3.533333333333335 1.4633333333333312z m-10.69-13.333333333333336l9.511666666666668 9.511666666666667c0.6333333333333329 0.629999999999999 1.7266666666666666 0.629999999999999 2.3583333333333343 0 0.31666666666666643-0.31666666666666643 0.4883333333333333-0.7333333333333343 0.4883333333333333-1.1799999999999997s-0.173333333333332-0.8633333333333333-0.4833333333333343-1.1766666666666659l-5.498333333333337-5.48833333333333h16.1c0.9200000000000017 0 1.6666666666666679-0.75 1.6666666666666679-1.6666666666666679s-0.7466666666666661-1.6666666666666679-1.6666666666666679-1.6666666666666679h-16.096666666666664l5.488333333333333-5.488333333333333c0.31666666666666643-0.31666666666666643 0.4883333333333333-0.7333333333333343 0.4883333333333333-1.1799999999999997s-0.173333333333332-0.8633333333333333-0.48666666666666814-1.1766666666666659c-0.6333333333333329-0.6333333333333329-1.7266666666666666-0.6333333333333329-2.3583333333333343 0l-9.510000000000002 9.511666666666667z' })
                )
            );
        }
    }]);

    return TiArrowLeftOutline;
}(React.Component);

exports.default = TiArrowLeftOutline;
module.exports = exports['default'];