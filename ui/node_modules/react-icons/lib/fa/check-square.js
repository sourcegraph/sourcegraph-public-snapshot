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

var FaCheckSquare = function (_React$Component) {
    _inherits(FaCheckSquare, _React$Component);

    function FaCheckSquare() {
        _classCallCheck(this, FaCheckSquare);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaCheckSquare).apply(this, arguments));
    }

    _createClass(FaCheckSquare, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm18.147142857142857 28.995714285714286l13.705714285714283-13.705714285714285q0.42428571428571615-0.4242857142857144 0.42428571428571615-1.0042857142857144t-0.4242857142857126-1.0042857142857144l-2.277142857142856-2.2771428571428576q-0.4242857142857126-0.4242857142857144-1.0042857142857144-0.4242857142857144t-1.0042857142857144 0.4242857142857144l-10.424285714285716 10.424285714285716-4.710000000000001-4.710000000000001q-0.4242857142857144-0.4242857142857126-1.0042857142857144-0.4242857142857126t-1.0042857142857144 0.4242857142857126l-2.2771428571428576 2.277142857142856q-0.4242857142857144 0.4242857142857126-0.4242857142857144 1.0042857142857144t0.4242857142857144 1.0042857142857144l7.991428571428573 7.991428571428571q0.4242857142857126 0.4242857142857126 1.0042857142857144 0.4242857142857126t1.0042857142857144-0.4242857142857126z m18.99571428571429-19.71v21.42857142857143q0 2.6571428571428584-1.8857142857142861 4.5428571428571445t-4.5428571428571445 1.8857142857142861h-21.42857142857143q-2.6571428571428575 0-4.542857142857144-1.8857142857142861t-1.885714285714284-4.5428571428571445v-21.42857142857143q0-2.6571428571428575 1.8857142857142857-4.542857142857144t4.542857142857144-1.885714285714284h21.42857142857143q2.6571428571428584 0 4.5428571428571445 1.8857142857142857t1.8857142857142861 4.542857142857144z' })
                )
            );
        }
    }]);

    return FaCheckSquare;
}(React.Component);

exports.default = FaCheckSquare;
module.exports = exports['default'];