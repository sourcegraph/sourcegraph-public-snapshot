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

var FaStickyNote = function (_React$Component) {
    _inherits(FaStickyNote, _React$Component);

    function FaStickyNote() {
        _classCallCheck(this, FaStickyNote);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaStickyNote).apply(this, arguments));
    }

    _createClass(FaStickyNote, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm25.714285714285715 27.857142857142858v9.285714285714288h-20.714285714285715q-0.8928571428571432 0-1.5171428571428573-0.6257142857142881t-0.6257142857142854-1.5171428571428578v-30q0-0.8928571428571432 0.6257142857142859-1.5171428571428573t1.517142857142857-0.6257142857142854h30q0.8928571428571459 0 1.5171428571428578 0.6257142857142859t0.6257142857142881 1.517142857142857v20.714285714285715h-9.285714285714285q-0.8928571428571423 0-1.5171428571428578 0.6257142857142846t-0.6257142857142881 1.5171428571428578z m2.8571428571428577 0.7142857142857153h8.504285714285711q-0.33428571428571274 1.8285714285714292-1.451428571428572 2.9471428571428575l-4.107142857142858 4.107142857142858q-1.1142857142857139 1.1142857142857139-2.9471428571428575 1.451428571428572v-8.505714285714287z' })
                )
            );
        }
    }]);

    return FaStickyNote;
}(React.Component);

exports.default = FaStickyNote;
module.exports = exports['default'];