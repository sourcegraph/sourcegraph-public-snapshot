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

var FaClone = function (_React$Component) {
    _inherits(FaClone, _React$Component);

    function FaClone() {
        _classCallCheck(this, FaClone);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaClone).apply(this, arguments));
    }

    _createClass(FaClone, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm37.142857142857146 36.42857142857143v-24.285714285714285q0-0.28999999999999915-0.21142857142856997-0.5028571428571436t-0.5028571428571453-0.21142857142857352h-24.285714285714285q-0.28999999999999915 0-0.5028571428571436 0.21142857142857174t-0.21142857142857352 0.5028571428571418v24.28571428571429q0 0.28999999999999915 0.21142857142857174 0.5028571428571453t0.5028571428571418 0.21142857142856997h24.28571428571429q0.28999999999999915 0 0.5028571428571453-0.21142857142856997t0.21142857142856997-0.5028571428571453z m2.857142857142854-24.285714285714285v24.285714285714285q0 1.471428571428575-1.048571428571428 2.5228571428571414t-2.5228571428571414 1.048571428571428h-24.285714285714285q-1.4714285714285715 0-2.522857142857143-1.048571428571428t-1.0485714285714316-2.5228571428571414v-24.285714285714285q0-1.4714285714285715 1.048571428571428-2.522857142857143t2.522857142857143-1.0485714285714316h24.28571428571429q1.471428571428575 0 2.5228571428571414 1.048571428571428t1.048571428571428 2.522857142857143z m-8.57142857142857-8.571428571428571v3.5714285714285685h-2.8571428571428577v-3.5714285714285716q0-0.29000000000000004-0.21142857142856997-0.5028571428571431t-0.5028571428571453-0.2114285714285713h-24.285714285714285q-0.29000000000000004 0-0.5028571428571427 0.2114285714285713t-0.21142857142857308 0.5028571428571431v24.285714285714285q0 0.28999999999999915 0.2114285714285713 0.5028571428571418t0.5028571428571431 0.21142857142857352h3.5714285714285716v2.8571428571428577h-3.5714285714285716q-1.4714285714285715 0-2.522857142857143-1.048571428571428t-1.0485714285714285-2.522857142857145v-24.285714285714285q0-1.4714285714285729 1.0485714285714285-2.5228571428571445t2.522857142857143-1.0485714285714285h24.285714285714285q1.4714285714285715 0 2.522857142857145 1.0485714285714285t1.048571428571428 2.522857142857143z' })
                )
            );
        }
    }]);

    return FaClone;
}(React.Component);

exports.default = FaClone;
module.exports = exports['default'];