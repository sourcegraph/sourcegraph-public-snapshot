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

var FaSpoon = function (_React$Component) {
    _inherits(FaSpoon, _React$Component);

    function FaSpoon() {
        _classCallCheck(this, FaSpoon);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaSpoon).apply(this, arguments));
    }

    _createClass(FaSpoon, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm27.142857142857142 11.785714285714286q0 3.2371428571428567-1.2714285714285722 5.435714285714285t-3.394285714285715 3.024285714285714l1.0042857142857144 18.325714285714284q0.04285714285714448 0.5799999999999983-0.35714285714285765 1.0042857142857144t-0.9814285714285695 0.42428571428571615h-4.285714285714285q-0.581428571428571 0-0.985714285714284-0.42428571428571615t-0.35714285714285765-1.0042857142857144l1.0057142857142871-18.325714285714284q-2.120000000000001-0.8257142857142838-3.3928571428571423-3.024285714285714t-1.2742857142857176-5.435714285714285q0-2.8571428571428577 0.9485714285714284-5.5685714285714285t2.6228571428571446-4.464285714285714 3.572857142857142-1.7528571428571436 3.571428571428573 1.7528571428571431 2.6228571428571428 4.4642857142857135 0.9485714285714302 5.568571428571428z' })
                )
            );
        }
    }]);

    return FaSpoon;
}(React.Component);

exports.default = FaSpoon;
module.exports = exports['default'];