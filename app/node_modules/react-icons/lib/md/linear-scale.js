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

var MdLinearScale = function (_React$Component) {
    _inherits(MdLinearScale, _React$Component);

    function MdLinearScale() {
        _classCallCheck(this, MdLinearScale);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdLinearScale).apply(this, arguments));
    }

    _createClass(MdLinearScale, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm32.5 15.860000000000001q1.7166666666666686 0 2.9299999999999997 1.2116666666666678t1.211666666666666 2.928333333333331-1.211666666666666 2.9333333333333336-2.9299999999999997 1.2100000000000009q-2.8133333333333326 0-3.828333333333333-2.5h-4.843333333333334q-1.0166666666666657 2.5-3.828333333333333 2.5t-3.828333333333333-2.5h-4.843333333333334q-1.0166666666666657 2.5-3.828333333333333 2.5-1.7166666666666668 0-2.9299999999999997-1.211666666666666t-1.2116666666666664-2.9316666666666684 1.2116666666666664-2.9283333333333346 2.9299999999999997-1.2116666666666642q2.8133333333333326 0 3.828333333333333 2.4999999999999982h4.843333333333334q1.0166666666666657-2.5 3.828333333333333-2.5t3.828333333333333 2.5h4.843333333333334q1.0166666666666657-2.5 3.828333333333333-2.5z' })
                )
            );
        }
    }]);

    return MdLinearScale;
}(React.Component);

exports.default = MdLinearScale;
module.exports = exports['default'];