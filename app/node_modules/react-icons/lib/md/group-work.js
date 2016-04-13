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

var MdGroupWork = function (_React$Component) {
    _inherits(MdGroupWork, _React$Component);

    function MdGroupWork() {
        _classCallCheck(this, MdGroupWork);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdGroupWork).apply(this, arguments));
    }

    _createClass(MdGroupWork, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm26.64 29.140000000000004q1.7166666666666686 0 2.9666666666666686-1.211666666666666t1.25-2.9299999999999997-1.25-2.9333333333333336-2.9666666666666686-1.2100000000000009-2.9299999999999997 1.211666666666666-1.2100000000000009 2.93333333333333 1.2100000000000009 2.9283333333333346 2.9299999999999997 1.211666666666666z m-10.78-15.780000000000003q0 1.7166666666666668 1.2116666666666678 2.9300000000000015t2.928333333333331 1.2099999999999973 2.9333333333333336-1.2100000000000009 1.2100000000000009-2.9299999999999997-1.211666666666666-2.966666666666667-2.9333333333333336-1.25-2.9283333333333346 1.25-1.211666666666666 2.966666666666667z m-2.5 15.78q1.7166666666666668 0 2.9300000000000015-1.211666666666666t1.2099999999999973-2.9283333333333346-1.2100000000000009-2.9333333333333336-2.9299999999999997-1.2100000000000009-2.966666666666667 1.211666666666666-1.25 2.9333333333333336 1.25 2.9283333333333346 2.966666666666667 1.211666666666666z m6.639999999999999-25.78q6.875 8.881784197001252e-16 11.758333333333333 4.883333333333335t4.883333333333333 11.756666666666666-4.883333333333333 11.759999999999998-11.758333333333333 4.88333333333334-11.758333333333333-4.883333333333333-4.883333333333333-11.760000000000005 4.883333333333333-11.756666666666668 11.758333333333333-4.883333333333332z' })
                )
            );
        }
    }]);

    return MdGroupWork;
}(React.Component);

exports.default = MdGroupWork;
module.exports = exports['default'];