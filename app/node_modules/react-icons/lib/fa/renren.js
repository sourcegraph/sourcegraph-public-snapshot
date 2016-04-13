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

var FaRenren = function (_React$Component) {
    _inherits(FaRenren, _React$Component);

    function FaRenren() {
        _classCallCheck(this, FaRenren);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaRenren).apply(this, arguments));
    }

    _createClass(FaRenren, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm28.147142857142857 35.042857142857144q-3.8185714285714276 2.1000000000000014-8.214285714285715 2.1000000000000014-4.375714285714283 0-8.19-2.1000000000000014 3.0785714285714274-1.9399999999999977 5.257142857142858-4.708571428571428t2.934285714285714-5.982857142857142q0.7814285714285703 3.2142857142857153 2.9571428571428555 5.982857142857142t5.257142857142856 4.709999999999997z m-11.047142857142855-31.871428571428574v10.824285714285715q0 5.625714285714288-2.8257142857142856 10.257142857142856t-7.377142857142857 6.842857142857145q-4.040000000000001-4.801428571428573-4.040000000000001-11.052857142857146-4.440892098500626e-16-4.174285714285714 1.8642857142857139-7.800000000000001t5.121428571428572-6.014285714285711 7.257142857142858-3.0571428571428574z m20.042857142857144 16.87142857142857q0 6.251428571428573-4.039999999999999 11.05-4.5528571428571425-2.210000000000001-7.377142857142857-6.842857142857142t-2.8242857142857147-10.257142857142856v-10.821428571428573q3.997142857142858 0.6714285714285722 7.257142857142856 3.0571428571428587t5.12142857142857 6.017142857142856 1.8628571428571448 7.797142857142855z' })
                )
            );
        }
    }]);

    return FaRenren;
}(React.Component);

exports.default = FaRenren;
module.exports = exports['default'];