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

var FaBookmark = function (_React$Component) {
    _inherits(FaBookmark, _React$Component);

    function FaBookmark() {
        _classCallCheck(this, FaBookmark);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaBookmark).apply(this, arguments));
    }

    _createClass(FaBookmark, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm31.697142857142858 2.857142857142857q0.5142857142857125 0 0.9828571428571422 0.20000000000000018 0.7371428571428567 0.29142857142857137 1.1714285714285708 0.9171428571428573t0.4357142857142833 1.3857142857142857v28.771428571428572q0 0.7571428571428598-0.43428571428571416 1.3857142857142861t-1.1714285714285708 0.914285714285711q-0.42428571428571615 0.1785714285714306-0.9828571428571422 0.1785714285714306-1.071428571428573 0-1.8528571428571432-0.7142857142857153l-9.845714285714283-9.464285714285715-9.842857142857143 9.464285714285715q-0.8042857142857134 0.7371428571428567-1.854285714285714 0.7371428571428567-0.5142857142857142 0-0.9828571428571431-0.20000000000000284-0.7371428571428575-0.29142857142856826-1.1714285714285717-0.9171428571428564t-0.4342857142857133-1.3828571428571337v-28.77285714285715q0-0.7599999999999998 0.43428571428571416-1.3857142857142857t1.1714285714285717-0.9142857142857141q0.468571428571428-0.2028571428571424 0.982857142857144-0.2028571428571424h23.392857142857146z' })
                )
            );
        }
    }]);

    return FaBookmark;
}(React.Component);

exports.default = FaBookmark;
module.exports = exports['default'];