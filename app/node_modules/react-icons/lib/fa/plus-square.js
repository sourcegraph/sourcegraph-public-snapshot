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

var FaPlusSquare = function (_React$Component) {
    _inherits(FaPlusSquare, _React$Component);

    function FaPlusSquare() {
        _classCallCheck(this, FaPlusSquare);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaPlusSquare).apply(this, arguments));
    }

    _createClass(FaPlusSquare, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm31.42857142857143 21.42857142857143v-2.8571428571428577q0-0.5800000000000018-0.4242857142857126-1.0042857142857144t-1.004285714285718-0.42428571428571615h-7.142857142857142v-7.142857142857142q0-0.5800000000000001-0.4242857142857126-1.0042857142857144t-1.0042857142857144-0.4242857142857144h-2.8571428571428577q-0.5800000000000018 0-1.0042857142857144 0.4242857142857144t-0.42428571428571615 1.0042857142857144v7.142857142857142h-7.142857142857142q-0.5800000000000001 0-1.0042857142857144 0.4242857142857126t-0.4242857142857144 1.004285714285718v2.8571428571428577q0 0.5799999999999983 0.4242857142857144 1.0042857142857144t1.0042857142857144 0.4242857142857126h7.142857142857142v7.142857142857142q0 0.5799999999999983 0.4242857142857126 1.0042857142857144t1.004285714285718 0.42428571428571615h2.8571428571428577q0.5799999999999983 0 1.0042857142857144-0.4242857142857126t0.4242857142857126-1.004285714285718v-7.142857142857142h7.142857142857142q0.5799999999999983 0 1.0042857142857144-0.4242857142857126t0.42428571428571615-1.0042857142857144z m5.714285714285715-12.142857142857142v21.42857142857143q0 2.657142857142855-1.8857142857142861 4.542857142857141t-4.5428571428571445 1.8857142857142861h-21.42857142857143q-2.6571428571428575 0-4.542857142857144-1.8857142857142861t-1.885714285714284-4.5428571428571445v-21.42857142857143q0-2.6571428571428575 1.8857142857142857-4.542857142857144t4.542857142857144-1.885714285714284h21.42857142857143q2.6571428571428584 0 4.5428571428571445 1.8857142857142857t1.8857142857142861 4.542857142857144z' })
                )
            );
        }
    }]);

    return FaPlusSquare;
}(React.Component);

exports.default = FaPlusSquare;
module.exports = exports['default'];