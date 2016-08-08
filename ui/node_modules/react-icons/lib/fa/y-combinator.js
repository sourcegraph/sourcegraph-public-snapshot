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

var FaYCombinator = function (_React$Component) {
    _inherits(FaYCombinator, _React$Component);

    function FaYCombinator() {
        _classCallCheck(this, FaYCombinator);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaYCombinator).apply(this, arguments));
    }

    _createClass(FaYCombinator, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm20.914285714285715 22.41l5.938571428571429-11.138571428571428h-2.5l-3.5042857142857144 6.964285714285715q-0.5357142857142847 1.071428571428573-0.9828571428571422 2.0528571428571425l-0.937142857142856-2.0528571428571425-3.4600000000000026-6.964285714285717h-2.678571428571429l5.871428571428574 11.004285714285716v7.232857142857142h2.252857142857142v-7.100000000000001z m16.22857142857143-19.552857142857142v34.28571428571428h-34.28571428571429v-34.285714285714285h34.285714285714285z' })
                )
            );
        }
    }]);

    return FaYCombinator;
}(React.Component);

exports.default = FaYCombinator;
module.exports = exports['default'];