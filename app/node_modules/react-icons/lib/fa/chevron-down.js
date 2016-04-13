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

var FaChevronDown = function (_React$Component) {
    _inherits(FaChevronDown, _React$Component);

    function FaChevronDown() {
        _classCallCheck(this, FaChevronDown);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaChevronDown).apply(this, arguments));
    }

    _createClass(FaChevronDown, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm37.56714285714286 18.035714285714285l-16.562857142857148 16.54q-0.4242857142857126 0.42428571428571615-1.0042857142857144 0.42428571428571615t-1.0042857142857144-0.42428571428571615l-16.562857142857144-16.54q-0.42428571428571304-0.4242857142857126-0.42428571428571304-1.0142857142857125t0.4242857142857144-1.0171428571428578l3.705714285714285-3.6828571428571433q0.4242857142857144-0.4242857142857144 1.0042857142857144-0.4242857142857144t1.0042857142857144 0.4242857142857144l11.852857142857143 11.852857142857145 11.852857142857143-11.852857142857143q0.4242857142857126-0.4242857142857144 1.004285714285711-0.4242857142857144t1.0042857142857144 0.4242857142857144l3.7057142857142864 3.6828571428571415q0.42428571428571615 0.4242857142857126 0.42428571428571615 1.0142857142857125t-0.42428571428571615 1.0171428571428578z' })
                )
            );
        }
    }]);

    return FaChevronDown;
}(React.Component);

exports.default = FaChevronDown;
module.exports = exports['default'];