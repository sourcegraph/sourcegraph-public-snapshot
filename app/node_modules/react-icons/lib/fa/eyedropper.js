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

var FaEyedropper = function (_React$Component) {
    _inherits(FaEyedropper, _React$Component);

    function FaEyedropper() {
        _classCallCheck(this, FaEyedropper);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaEyedropper).apply(this, arguments));
    }

    _createClass(FaEyedropper, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm37.900000000000006 2.1q2.0999999999999943 2.0957142857142856 2.0999999999999943 5.0528571428571425t-2.0999999999999943 5.032857142857143l-5.021428571428572 4.9785714285714295 2.3214285714285694 2.321428571428573q0.2228571428571442 0.2228571428571442 0.2228571428571442 0.5142857142857125t-0.2228571428571442 0.5114285714285707l-4.685714285714287 4.685714285714287q-0.2242857142857133 0.2242857142857133-0.514285714285716 0.2242857142857133t-0.5142857142857125-0.2228571428571442l-2.342857142857145-2.3457142857142834-13.46 13.461428571428574q-0.8257142857142856 0.8257142857142838-2.008571428571429 0.8257142857142838h-4.531428571428569l-5.714285714285714 2.857142857142854-1.4285714285714286-1.4285714285714306 2.8571428571428577-5.714285714285715v-4.53142857142857q0-1.1828571428571415 0.8257142857142856-2.008571428571429l13.459999999999999-13.459999999999996-2.3428571428571434-2.3428571428571434q-0.2242857142857151-0.2242857142857151-0.2242857142857151-0.5142857142857142t0.22285714285714242-0.5142857142857142l4.685714285714285-4.685714285714286q0.2228571428571442-0.2228571428571433 0.5142857142857125-0.2228571428571433t0.5114285714285707 0.22285714285714242l2.321428571428573 2.321428571428571 4.977142857142859-5.022857142857143q2.0757142857142874-2.1 5.032857142857139-2.1t5.057142857142857 2.1z m-26.471428571428575 30.75714285714286l12.857142857142854-12.857142857142858-4.285714285714285-4.285714285714285-12.857142857142858 12.857142857142858v4.285714285714285h4.285714285714285z' })
                )
            );
        }
    }]);

    return FaEyedropper;
}(React.Component);

exports.default = FaEyedropper;
module.exports = exports['default'];