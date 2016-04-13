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

var FaArchive = function (_React$Component) {
    _inherits(FaArchive, _React$Component);

    function FaArchive() {
        _classCallCheck(this, FaArchive);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaArchive).apply(this, arguments));
    }

    _createClass(FaArchive, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm24.285714285714285 18.571428571428573q0-0.5800000000000018-0.4242857142857126-1.0042857142857144t-1.0042857142857144-0.42428571428571615h-5.714285714285715q-0.5800000000000018 0-1.0042857142857144 0.4242857142857126t-0.4242857142857126 1.004285714285718 0.4242857142857126 1.0042857142857144 1.0042857142857144 0.4242857142857126h5.714285714285715q0.5800000000000018 0 1.0042857142857144-0.4242857142857126t0.4242857142857126-1.0042857142857144z m12.857142857142854-4.285714285714285v21.428571428571434q0 0.5799999999999983-0.42428571428571615 1.0042857142857144t-1.0042857142857073 0.42428571428570905h-31.42857142857143q-0.5799999999999992 0-1.0042857142857136-0.42428571428571615t-0.42428571428571393-1.0042857142857144v-21.42857142857143q0-0.5799999999999983 0.4242857142857144-1.0042857142857127t1.004285714285714-0.4242857142857144h31.42857142857143q0.5799999999999983 0 1.0042857142857144 0.4242857142857144t0.42428571428571615 1.0042857142857144z m1.4285714285714306-10v5.714285714285715q0 0.5800000000000001-0.42428571428571615 1.0042857142857144t-1.0042857142857073 0.4242857142857108h-34.28571428571429q-0.579999999999997 0-1.0042857142857111-0.4242857142857144t-0.4242857142857144-1.0042857142857144v-5.714285714285714q0-0.5800000000000001 0.42428571428571416-1.0042857142857144t1.0042857142857144-0.42428571428571393h34.285714285714285q0.5799999999999983 0 1.0042857142857144 0.4242857142857144t0.42428571428571615 1.004285714285714z' })
                )
            );
        }
    }]);

    return FaArchive;
}(React.Component);

exports.default = FaArchive;
module.exports = exports['default'];