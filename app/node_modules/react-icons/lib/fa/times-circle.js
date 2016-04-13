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

var FaTimesCircle = function (_React$Component) {
    _inherits(FaTimesCircle, _React$Component);

    function FaTimesCircle() {
        _classCallCheck(this, FaTimesCircle);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaTimesCircle).apply(this, arguments));
    }

    _createClass(FaTimesCircle, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm28.504285714285714 25.042857142857144q0-0.5785714285714292-0.4242857142857126-1.0028571428571418l-4.040000000000003-4.040000000000003 4.039999999999999-4.039999999999999q0.4242857142857126-0.4242857142857144 0.4242857142857126-1.0042857142857144 0-0.6042857142857141-0.4242857142857126-1.0285714285714285l-2.0085714285714253-2.007142857142858q-0.4242857142857126-0.4242857142857144-1.0285714285714285-0.4242857142857144-0.5785714285714292 0-1.0028571428571418 0.4242857142857144l-4.040000000000003 4.040000000000001-4.039999999999999-4.040000000000001q-0.4242857142857144-0.4242857142857144-1.0042857142857144-0.4242857142857144-0.6042857142857141 0-1.0285714285714285 0.4242857142857144l-2.0071428571428562 2.008571428571429q-0.4242857142857144 0.4242857142857144-0.4242857142857144 1.0285714285714285 0 0.5785714285714292 0.4242857142857144 1.0028571428571436l4.039999999999999 4.039999999999999-4.039999999999999 4.039999999999999q-0.4242857142857144 0.4242857142857126-0.4242857142857144 1.0042857142857144 0 0.6042857142857159 0.4242857142857144 1.0285714285714285l2.008571428571429 2.008571428571429q0.4242857142857144 0.4242857142857126 1.0285714285714285 0.4242857142857126 0.5785714285714292 0 1.0028571428571436-0.4242857142857126l4.039999999999997-4.041428571428572 4.039999999999999 4.039999999999999q0.4242857142857126 0.4242857142857126 1.0042857142857144 0.4242857142857126 0.6042857142857159 0 1.0285714285714285-0.4242857142857126l2.008571428571429-2.008571428571429q0.4242857142857126-0.4242857142857126 0.4242857142857126-1.0285714285714285z m8.638571428571431-5.0428571428571445q0 4.665714285714287-2.299999999999997 8.604285714285716t-6.237142857142857 6.238571428571426-8.605714285714292 2.3000000000000043-8.6-2.3000000000000043-6.242857142857143-6.238571428571426-2.295714285714286-8.604285714285716 2.3000000000000003-8.604285714285714 6.234285714285714-6.238571428571428 8.604285714285714-2.3000000000000003 8.605714285714285 2.3000000000000003 6.238571428571426 6.238571428571428 2.298571428571435 8.604285714285714z' })
                )
            );
        }
    }]);

    return FaTimesCircle;
}(React.Component);

exports.default = FaTimesCircle;
module.exports = exports['default'];