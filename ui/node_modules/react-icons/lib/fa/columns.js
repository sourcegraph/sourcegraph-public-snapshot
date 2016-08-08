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

var FaColumns = function (_React$Component) {
    _inherits(FaColumns, _React$Component);

    function FaColumns() {
        _classCallCheck(this, FaColumns);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaColumns).apply(this, arguments));
    }

    _createClass(FaColumns, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm5 34.285714285714285h13.571428571428573v-25.714285714285715h-14.285714285714288v25q8.881784197001252e-16 0.28999999999999915 0.21142857142857263 0.5028571428571453t0.5028571428571427 0.21142857142856997z m30.714285714285715-0.7142857142857153v-25h-14.285714285714285v25.714285714285715h13.57142857142857q0.28999999999999915 0 0.5028571428571453-0.21142857142856997t0.21142857142856997-0.5028571428571453z m2.857142857142854-27.142857142857142v27.142857142857142q0 1.471428571428575-1.048571428571428 2.5228571428571414t-2.5228571428571414 1.0485714285714351h-30q-1.4714285714285715 0-2.522857142857143-1.048571428571428t-1.0485714285714283-2.5228571428571485v-27.142857142857142q0-1.4714285714285715 1.0485714285714283-2.522857142857143t2.522857142857143-1.0485714285714267h30q1.471428571428575 0 2.5228571428571414 1.0485714285714285t1.048571428571428 2.522857142857143z' })
                )
            );
        }
    }]);

    return FaColumns;
}(React.Component);

exports.default = FaColumns;
module.exports = exports['default'];