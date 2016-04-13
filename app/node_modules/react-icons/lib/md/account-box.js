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

var MdAccountBox = function (_React$Component) {
    _inherits(MdAccountBox, _React$Component);

    function MdAccountBox() {
        _classCallCheck(this, MdAccountBox);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdAccountBox).apply(this, arguments));
    }

    _createClass(MdAccountBox, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm10 28.36v1.6400000000000006h20v-1.6400000000000006q0-2.2666666666666657-3.4383333333333326-3.711666666666666t-6.561666666666667-1.4483333333333341-6.566666666666666 1.4450000000000003-3.4333333333333336 3.7133333333333347z m15-13.36q0-2.033333333333333-1.4833333333333343-3.5166666666666657t-3.5166666666666657-1.4833333333333343-3.5166666666666657 1.4833333333333343-1.4833333333333343 3.5166666666666657 1.4833333333333343 3.5166666666666657 3.5166666666666657 1.4833333333333343 3.5166666666666657-1.4833333333333343 1.4833333333333343-3.5166666666666657z m-20-6.639999999999999q0-1.3283333333333331 0.9766666666666666-2.3433333333333337t2.3833333333333346-1.0166666666666675h23.28333333333334q1.326666666666668 0 2.3416666666666686 1.0166666666666666t1.0149999999999935 2.3433333333333346v23.28333333333334q0 1.326666666666668-1.0166666666666657 2.3416666666666686t-2.3416666666666686 1.0166666666666657h-23.28333333333333q-1.405000000000002 0-2.3833333333333346-1.0166666666666657t-0.9749999999999996-2.341666666666672v-23.285000000000004z' })
                )
            );
        }
    }]);

    return MdAccountBox;
}(React.Component);

exports.default = MdAccountBox;
module.exports = exports['default'];