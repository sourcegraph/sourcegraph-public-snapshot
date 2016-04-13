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

var FaTrello = function (_React$Component) {
    _inherits(FaTrello, _React$Component);

    function FaTrello() {
        _classCallCheck(this, FaTrello);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaTrello).apply(this, arguments));
    }

    _createClass(FaTrello, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm18.571428571428573 30v-22.857142857142858q0-0.31428571428571317-0.1999999999999993-0.5142857142857133t-0.5142857142857125-0.20000000000000018h-10.714285714285719q-0.31428571428571317 0-0.5142857142857133 0.20000000000000018t-0.20000000000000018 0.5142857142857142v22.857142857142858q0 0.31428571428571317 0.20000000000000018 0.5142857142857125t0.5142857142857142 0.1999999999999993h10.714285714285715q0.31428571428571317 0 0.514285714285716-0.1999999999999993t0.1999999999999993-0.5142857142857125z m14.999999999999996-8.57142857142857v-14.285714285714288q0-0.31428571428571317-0.20000000000000284-0.5142857142857133t-0.5142857142857054-0.20000000000000018h-10.714285714285715q-0.31428571428571317 0-0.5142857142857125 0.20000000000000018t-0.20000000000000284 0.5142857142857142v14.285714285714288q0 0.31428571428571317 0.1999999999999993 0.5142857142857125t0.514285714285716 0.1999999999999993h10.714285714285715q0.3142857142857167 0 0.5142857142857125-0.1999999999999993t0.20000000000000284-0.5142857142857125z m3.5714285714285765-17.142857142857146v31.42857142857143q0 0.5799999999999983-0.42428571428571615 1.0042857142857144t-1.0042857142857144 0.42428571428571615h-31.42857142857143q-0.5799999999999992 0-1.0042857142857136-0.42428571428571615t-0.42428571428571393-1.0042857142857144v-31.42857142857143q0-0.5799999999999992 0.4242857142857144-1.0042857142857136t1.004285714285714-0.42428571428571393h31.42857142857143q0.5799999999999983 0 1.0042857142857144 0.4242857142857144t0.42428571428571615 1.004285714285714z' })
                )
            );
        }
    }]);

    return FaTrello;
}(React.Component);

exports.default = FaTrello;
module.exports = exports['default'];