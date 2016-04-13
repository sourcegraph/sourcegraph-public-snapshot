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

var FaBlackTie = function (_React$Component) {
    _inherits(FaBlackTie, _React$Component);

    function FaBlackTie() {
        _classCallCheck(this, FaBlackTie);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaBlackTie).apply(this, arguments));
    }

    _createClass(FaBlackTie, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm2.857142857142857 2.857142857142857h34.285714285714285v34.285714285714285h-34.285714285714285v-34.285714285714285z m24.21857142857143 24.88857142857143l-4.932857142857145-14.085714285714289 4.9328571428571415-6.628571428571428h-14.151428571428568l4.9328571428571415 6.628571428571428-4.9328571428571415 14.085714285714285 7.075714285714284 6.785714285714285z' })
                )
            );
        }
    }]);

    return FaBlackTie;
}(React.Component);

exports.default = FaBlackTie;
module.exports = exports['default'];