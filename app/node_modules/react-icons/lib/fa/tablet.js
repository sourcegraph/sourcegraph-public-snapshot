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

var FaTablet = function (_React$Component) {
    _inherits(FaTablet, _React$Component);

    function FaTablet() {
        _classCallCheck(this, FaTablet);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaTablet).apply(this, arguments));
    }

    _createClass(FaTablet, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm21.42857142857143 31.42857142857143q0-0.5799999999999983-0.4242857142857126-1.0042857142857144t-1.004285714285718-0.42428571428571615-1.0042857142857144 0.4242857142857126-0.4242857142857126 1.004285714285718 0.4242857142857126 1.0042857142857144 1.0042857142857144 0.42428571428571615 1.0042857142857144-0.42428571428571615 0.42428571428571615-1.0042857142857144z m8.57142857142857-3.571428571428573v-21.42857142857143q0-0.29000000000000004-0.21142857142856997-0.5028571428571427t-0.5028571428571453-0.21142857142856997h-18.571428571428573q-0.28999999999999915 0-0.5028571428571436 0.21142857142857174t-0.2114285714285682 0.5028571428571427v21.42857142857143q0 0.28999999999999915 0.21142857142857174 0.5028571428571418t0.5028571428571436 0.21142857142856997h18.571428571428573q0.28999999999999915 0 0.5028571428571418-0.21142857142856997t0.21142857142856997-0.5028571428571453z m2.857142857142854-21.42857142857143v24.285714285714285q0 1.4714285714285715-1.048571428571428 2.522857142857145t-2.5228571428571414 1.048571428571428h-18.571428571428573q-1.4714285714285715 0-2.522857142857143-1.048571428571428t-1.0485714285714254-2.5228571428571414v-24.285714285714285q0-1.4714285714285715 1.048571428571429-2.522857142857143t2.522857142857143-1.0485714285714303h18.571428571428573q1.4714285714285715 0 2.522857142857145 1.0485714285714285t1.048571428571428 2.522857142857143z' })
                )
            );
        }
    }]);

    return FaTablet;
}(React.Component);

exports.default = FaTablet;
module.exports = exports['default'];