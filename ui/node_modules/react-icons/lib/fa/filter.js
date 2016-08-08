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

var FaFilter = function (_React$Component) {
    _inherits(FaFilter, _React$Component);

    function FaFilter() {
        _classCallCheck(this, FaFilter);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaFilter).apply(this, arguments));
    }

    _createClass(FaFilter, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm35.60285714285715 6.585714285714286q0.38000000000000256 0.9142857142857146-0.3142857142857167 1.5614285714285705l-11.002857142857145 11.004285714285714v16.562857142857144q0 0.9371428571428595-0.8714285714285701 1.317142857142855-0.28571428571428825 0.11142857142857565-0.5571428571428569 0.11142857142857565-0.6028571428571432 0-1.0042857142857144-0.42428571428571615l-5.714285714285715-5.714285714285715q-0.4242857142857126-0.42428571428571615-0.4242857142857126-1.0042857142857144v-10.848571428571429l-11.004285714285714-11.004285714285713q-0.6914285714285722-0.6471428571428586-0.31428571428571495-1.5614285714285723 0.38285714285714256-0.8714285714285719 1.3185714285714285-0.8714285714285719h28.57142857142857q0.9371428571428595 0 1.317142857142855 0.871428571428571z' })
                )
            );
        }
    }]);

    return FaFilter;
}(React.Component);

exports.default = FaFilter;
module.exports = exports['default'];