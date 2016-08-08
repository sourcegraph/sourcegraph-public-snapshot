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

var FaCheck = function (_React$Component) {
    _inherits(FaCheck, _React$Component);

    function FaCheck() {
        _classCallCheck(this, FaCheck);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaCheck).apply(this, arguments));
    }

    _createClass(FaCheck, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm37.3 12.634285714285713q0 0.8928571428571423-0.6285714285714263 1.5171428571428578l-19.194285714285712 19.19714285714286q-0.6257142857142846 0.6257142857142881-1.5171428571428578 0.6257142857142881t-1.5142857142857142-0.6257142857142881l-11.117142857142857-11.114285714285714q-0.6257142857142859-0.6285714285714299-0.6257142857142859-1.518571428571427t0.6257142857142859-1.514285714285716l3.0357142857142856-3.0371428571428574q0.6257142857142854-0.6257142857142863 1.517142857142857-0.6257142857142863t1.5171428571428578 0.6257142857142863l6.562857142857144 6.585714285714285 14.638571428571428-14.670000000000003q0.6285714285714299-0.6257142857142854 1.518571428571427-0.6257142857142854t1.5171428571428578 0.6242857142857146l3.0357142857142847 3.0357142857142847q0.6257142857142881 0.6242857142857137 0.6257142857142881 1.5171428571428578z' })
                )
            );
        }
    }]);

    return FaCheck;
}(React.Component);

exports.default = FaCheck;
module.exports = exports['default'];