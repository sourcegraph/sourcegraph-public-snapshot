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

var FaUnlock = function (_React$Component) {
    _inherits(FaUnlock, _React$Component);

    function FaUnlock() {
        _classCallCheck(this, FaUnlock);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaUnlock).apply(this, arguments));
    }

    _createClass(FaUnlock, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm38.57142857142857 12.857142857142858v5.714285714285715q0 0.5800000000000018-0.42428571428571615 1.0042857142857144t-1.0042857142857073 0.4242857142857126h-1.4285714285714306q-0.5799999999999983 0-1.0042857142857144-0.4242857142857126t-0.42428571428571615-1.0042857142857144v-5.714285714285715q0-2.3657142857142865-1.6742857142857162-4.039999999999999t-4.039999999999996-1.6742857142857153-4.039999999999999 1.6742857142857135-1.6742857142857162 4.040000000000001v4.285714285714285h2.1428571428571423q0.8928571428571423 0 1.5171428571428578 0.6257142857142846t0.6257142857142846 1.5171428571428578v12.857142857142854q0 0.8928571428571459-0.6257142857142846 1.5171428571428578t-1.5171428571428578 0.6257142857142881h-21.42857142857143q-0.8928571428571428 0-1.5171428571428573-0.6257142857142881t-0.6257142857142834-1.5171428571428507v-12.857142857142858q0-0.8928571428571423 0.6257142857142857-1.5171428571428578t1.5171428571428573-0.6257142857142881h15.000000000000002v-4.285714285714285q0-4.12857142857143 2.935714285714287-7.064285714285715t7.064285714285713-2.9357142857142855 7.064285714285713 2.9357142857142855 2.9357142857142833 7.064285714285715z' })
                )
            );
        }
    }]);

    return FaUnlock;
}(React.Component);

exports.default = FaUnlock;
module.exports = exports['default'];