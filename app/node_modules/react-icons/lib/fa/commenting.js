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

var FaCommenting = function (_React$Component) {
    _inherits(FaCommenting, _React$Component);

    function FaCommenting() {
        _classCallCheck(this, FaCommenting);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaCommenting).apply(this, arguments));
    }

    _createClass(FaCommenting, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm14.285714285714286 20q0-1.1828571428571415-0.8371428571428563-2.0199999999999996t-2.0200000000000014-0.8371428571428581-2.0199999999999996 0.8371428571428581-0.8371428571428581 2.0199999999999996 0.8371428571428563 2.0199999999999996 2.0200000000000014 0.8371428571428581 2.0199999999999996-0.8371428571428581 0.8371428571428581-2.0199999999999996z m8.571428571428571 0q0-1.1828571428571415-0.8371428571428581-2.0199999999999996t-2.0199999999999996-0.8371428571428581-2.0199999999999996 0.8371428571428581-0.8371428571428581 2.0199999999999996 0.8371428571428581 2.0199999999999996 2.0199999999999996 0.8371428571428581 2.0199999999999996-0.8371428571428581 0.8371428571428581-2.0199999999999996z m8.571428571428573 0q0-1.1828571428571415-0.8371428571428581-2.0199999999999996t-2.0199999999999996-0.8371428571428581-2.0199999999999996 0.8371428571428581-0.8371428571428581 2.0199999999999996 0.8371428571428581 2.0199999999999996 2.0199999999999996 0.8371428571428581 2.0199999999999996-0.8371428571428581 0.8371428571428581-2.0199999999999996z m8.57142857142857 0q0 3.885714285714286-2.6785714285714306 7.175714285714285t-7.277142857142856 5.200000000000003-10.044285714285714 1.9099999999999966q-2.4571428571428555 0-4.710000000000001-0.3999999999999986-3.8614285714285703 3.857142857142861-9.709999999999999 5.10857142857143-1.1600000000000001 0.2228571428571442-1.92 0.28999999999999915-0.2671428571428569 0.022857142857141355-0.49142857142857155-0.134285714285717t-0.29000000000000004-0.3999999999999986q-0.08999999999999986-0.33571428571428896 0.44714285714285706-0.8285714285714292 0.11142857142857121-0.10999999999999943 0.524285714285714-0.47857142857142776t0.5685714285714285-0.5228571428571414 0.524285714285714-0.5685714285714312 0.5357142857142856-0.7028571428571411 0.4571428571428573-0.8257142857142838 0.4471428571428575-1.0714285714285694 0.32428571428571473-1.2828571428571394 0.27857142857142847-1.6185714285714283q-3.257142857142857-2.008571428571429-5.122857142857143-4.832857142857144t-1.862857142857143-6.01428571428572q0-3.885714285714286 2.6771428571428575-7.177142857142858t7.277142857142858-5.2 10.042857142857143-1.9114285714285728 10.045714285714286 1.9100000000000001 7.277142857142856 5.2 2.6799999999999997 7.178571428571431z' })
                )
            );
        }
    }]);

    return FaCommenting;
}(React.Component);

exports.default = FaCommenting;
module.exports = exports['default'];