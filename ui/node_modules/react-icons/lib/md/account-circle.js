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

var MdAccountCircle = function (_React$Component) {
    _inherits(MdAccountCircle, _React$Component);

    function MdAccountCircle() {
        _classCallCheck(this, MdAccountCircle);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdAccountCircle).apply(this, arguments));
    }

    _createClass(MdAccountCircle, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm20 32.03333333333333q6.328333333333333 0 10-5.391666666666666-0.07833333333333314-2.1883333333333326-3.5166666666666657-3.671666666666667t-6.483333333333334-1.4833333333333343-6.483333333333334 1.4433333333333351-3.5166666666666657 3.711666666666666q3.671666666666667 5.390000000000001 10 5.390000000000001z m0-23.673333333333336q-2.0333333333333314 0-3.5166666666666657 1.4833333333333343t-1.4833333333333343 3.518333333333336 1.4833333333333343 3.5166666666666675 3.5166666666666657 1.4833333333333343 3.5166666666666657-1.4833333333333343 1.4833333333333343-3.5166666666666657-1.4833333333333343-3.5166666666666657-3.5166666666666657-1.4833333333333343z m0-5q6.875 0 11.758333333333333 4.883333333333333t4.883333333333333 11.756666666666671-4.883333333333333 11.759999999999998-11.758333333333333 4.88333333333334-11.758333333333333-4.883333333333333-4.883333333333333-11.760000000000005 4.883333333333333-11.756666666666668 11.758333333333333-4.883333333333332z' })
                )
            );
        }
    }]);

    return MdAccountCircle;
}(React.Component);

exports.default = MdAccountCircle;
module.exports = exports['default'];