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

var FaLock = function (_React$Component) {
    _inherits(FaLock, _React$Component);

    function FaLock() {
        _classCallCheck(this, FaLock);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaLock).apply(this, arguments));
    }

    _createClass(FaLock, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm14.285714285714286 17.142857142857142h11.428571428571429v-4.285714285714285q0-2.3657142857142865-1.6742857142857126-4.039999999999999t-4.040000000000003-1.6742857142857153-4.039999999999999 1.6742857142857135-1.6742857142857144 4.040000000000001v4.285714285714285z m18.571428571428577 2.1428571428571423v12.857142857142854q0 0.8928571428571459-0.6257142857142881 1.5171428571428578t-1.5171428571428578 0.6257142857142881h-21.42857142857143q-0.8928571428571423 0-1.5171428571428578-0.6257142857142881t-0.6257142857142837-1.5171428571428507v-12.857142857142858q0-0.8928571428571423 0.6257142857142854-1.5171428571428578t1.5171428571428578-0.6257142857142881h0.7142857142857135v-4.285714285714285q0-4.107142857142858 2.9471428571428575-7.052857142857143t7.0528571428571425-2.947142857142857 7.0528571428571425 2.947142857142858 2.9471428571428575 7.0528571428571425v4.285714285714285h0.7142857142857153q0.8928571428571423 0 1.5171428571428578 0.6257142857142846t0.6257142857142881 1.5171428571428578z' })
                )
            );
        }
    }]);

    return FaLock;
}(React.Component);

exports.default = FaLock;
module.exports = exports['default'];