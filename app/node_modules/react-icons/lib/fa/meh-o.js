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

var FaMehO = function (_React$Component) {
    _inherits(FaMehO, _React$Component);

    function FaMehO() {
        _classCallCheck(this, FaMehO);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaMehO).apply(this, arguments));
    }

    _createClass(FaMehO, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm28.571428571428573 24.285714285714285q0 0.5799999999999983-0.4242857142857126 1.0042857142857144t-1.004285714285718 0.42428571428571615h-14.285714285714285q-0.5800000000000001 0-1.0042857142857144-0.4242857142857126t-0.4242857142857144-1.004285714285718 0.4242857142857144-1.0042857142857144 1.0042857142857144-0.4242857142857126h14.285714285714288q0.5799999999999983 0 1.0042857142857144 0.4242857142857126t0.4242857142857126 1.0042857142857144z m-11.428571428571427-10q0 1.1828571428571433-0.8371428571428581 2.0199999999999996t-2.0200000000000014 0.8371428571428581-2.0199999999999996-0.8371428571428581-0.8371428571428581-2.019999999999998 0.8371428571428563-2.0199999999999996 2.0200000000000014-0.8371428571428581 2.0200000000000014 0.8371428571428563 0.8371428571428545 2.0200000000000014z m11.42857142857143 0q0 1.1828571428571433-0.8371428571428581 2.0199999999999996t-2.020000000000003 0.8371428571428581-2.0199999999999996-0.8371428571428581-0.8371428571428581-2.019999999999998 0.8371428571428581-2.0199999999999996 2.0199999999999996-0.8371428571428581 2.0199999999999996 0.8371428571428563 0.8371428571428581 2.0200000000000014z m5.714285714285715 5.714285714285715q0-2.8999999999999986-1.1385714285714315-5.547142857142857t-3.047142857142859-4.5528571428571425-4.5528571428571425-3.047142857142857-5.547142857142859-1.1385714285714288-5.547142857142857 1.1385714285714288-4.5528571428571425 3.047142857142857-3.047142857142857 4.5528571428571425-1.1385714285714288 5.547142857142857 1.1385714285714288 5.547142857142859 3.047142857142857 4.5528571428571425 4.5528571428571425 3.047142857142859 5.547142857142857 1.1385714285714243 5.547142857142859-1.1385714285714315 4.5528571428571425-3.047142857142859 3.047142857142859-4.5528571428571425 1.1385714285714243-5.547142857142852z m2.857142857142854 0q0 4.665714285714287-2.299999999999997 8.604285714285716t-6.237142857142857 6.238571428571426-8.605714285714292 2.3000000000000043-8.6-2.3000000000000043-6.242857142857143-6.238571428571426-2.295714285714286-8.604285714285716 2.3000000000000003-8.604285714285714 6.234285714285714-6.238571428571428 8.604285714285714-2.3000000000000003 8.605714285714285 2.3000000000000003 6.238571428571426 6.238571428571428 2.298571428571435 8.604285714285714z' })
                )
            );
        }
    }]);

    return FaMehO;
}(React.Component);

exports.default = FaMehO;
module.exports = exports['default'];