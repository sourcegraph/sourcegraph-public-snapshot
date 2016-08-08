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

var FaBuysellads = function (_React$Component) {
    _inherits(FaBuysellads, _React$Component);

    function FaBuysellads() {
        _classCallCheck(this, FaBuysellads);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaBuysellads).apply(this, arguments));
    }

    _createClass(FaBuysellads, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm23.281428571428574 24.24285714285714h-6.562857142857144l3.2814285714285703-12.299999999999997z m1.918571428571429 7.18571428571429h6.942857142857143l-7.232857142857149-22.85714285714286h-9.82142857142857l-7.2314285714285695 22.85714285714286h6.9399999999999995l8.54857142857143-7.008571428571429z m11.942857142857143-22.142857142857146v21.42857142857143q0 2.634285714285717-1.8971428571428604 4.53142857142857t-4.53142857142857 1.8971428571428604h-21.42857142857143q-2.6342857142857143 0-4.531428571428572-1.8971428571428604t-1.8971428571428555-4.53142857142857v-21.42857142857143q0-2.6342857142857143 1.8971428571428572-4.531428571428572t4.531428571428572-1.8971428571428555h21.42857142857143q2.634285714285717 0 4.53142857142857 1.8971428571428572t1.8971428571428604 4.531428571428572z' })
                )
            );
        }
    }]);

    return FaBuysellads;
}(React.Component);

exports.default = FaBuysellads;
module.exports = exports['default'];