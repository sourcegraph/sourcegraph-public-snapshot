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

var FaMapMarker = function (_React$Component) {
    _inherits(FaMapMarker, _React$Component);

    function FaMapMarker() {
        _classCallCheck(this, FaMapMarker);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaMapMarker).apply(this, arguments));
    }

    _createClass(FaMapMarker, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm25.714285714285715 14.285714285714286q0-2.3657142857142848-1.6742857142857126-4.040000000000001t-4.040000000000003-1.6742857142857144-4.039999999999999 1.6742857142857144-1.6742857142857144 4.040000000000001 1.6742857142857144 4.040000000000001 4.039999999999999 1.6742857142857126 4.039999999999999-1.6742857142857126 1.6742857142857162-4.040000000000001z m5.714285714285715 0q0 2.4328571428571433-0.7371428571428567 3.9957142857142873l-8.125714285714288 17.275714285714283q-0.35714285714285765 0.7385714285714258-1.0599999999999987 1.1614285714285728t-1.5057142857142871 0.42428571428571615-1.508571428571429-0.42428571428571615-1.03857142857143-1.1600000000000037l-8.147142857142855-17.275714285714283q-0.7371428571428567-1.5642857142857132-0.7371428571428567-3.9971428571428564 0-4.732857142857144 3.3485714285714288-8.08t8.08-3.348571428571429 8.080000000000002 3.3485714285714283 3.3485714285714216 8.080000000000002z' })
                )
            );
        }
    }]);

    return FaMapMarker;
}(React.Component);

exports.default = FaMapMarker;
module.exports = exports['default'];