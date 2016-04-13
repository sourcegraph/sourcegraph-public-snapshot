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

var MdLocalDrink = function (_React$Component) {
    _inherits(MdLocalDrink, _React$Component);

    function MdLocalDrink() {
        _classCallCheck(this, MdLocalDrink);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdLocalDrink).apply(this, arguments));
    }

    _createClass(MdLocalDrink, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm30.546666666666667 13.360000000000001l0.7033333333333331-6.716666666666668h-22.5l0.7033333333333331 6.716666666666668h21.093333333333334z m-10.546666666666667 18.28q2.0333333333333314 3.552713678800501e-15 3.5166666666666657-1.4833333333333307t1.4833333333333343-3.5166666666666657q0-1.4833333333333343-1.25-3.75t-2.5-3.75l-1.25-1.4833333333333343q-5 5.623333333333335-5 8.98333333333333 0 2.0333333333333314 1.4833333333333343 3.5166666666666657t3.5166666666666657 1.4833333333333343z m-15-28.28h30l-3.359999999999996 30.39q-0.1566666666666663 1.25-1.0933333333333337 2.0700000000000003t-2.1883333333333326 0.8200000000000003h-16.71666666666667q-1.2500000000000018 0-2.1900000000000013-0.8200000000000003t-1.0916666666666668-2.0700000000000003z' })
                )
            );
        }
    }]);

    return MdLocalDrink;
}(React.Component);

exports.default = MdLocalDrink;
module.exports = exports['default'];