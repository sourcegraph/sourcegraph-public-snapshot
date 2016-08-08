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

var MdRoundedCorner = function (_React$Component) {
    _inherits(MdRoundedCorner, _React$Component);

    function MdRoundedCorner() {
        _classCallCheck(this, MdRoundedCorner);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdRoundedCorner).apply(this, arguments));
    }

    _createClass(MdRoundedCorner, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm35 13.360000000000001v8.283333333333333h-3.3599999999999994v-8.283333333333333q0-2.033333333333333-1.4833333333333343-3.5166666666666675t-3.5166666666666657-1.4833333333333343h-8.283333333333331v-3.3599999999999994h8.283333333333331q3.4383333333333326 0 5.899999999999999 2.461666666666667t2.460000000000001 5.8999999999999995z m-30 21.64v-3.3599999999999994h3.3599999999999994v3.3599999999999994h-3.3599999999999994z m6.640000000000001 0v-3.3599999999999994h3.3599999999999994v3.3599999999999994h-3.3599999999999994z m6.719999999999999 0v-3.3599999999999994h3.2833333333333314v3.3599999999999994h-3.2833333333333314z m-6.719999999999999-26.64v-3.3599999999999994h3.3599999999999994v3.3599999999999994h-3.3599999999999994z m-6.640000000000001 1.7763568394002505e-15v-3.360000000000001h3.3599999999999994v3.3599999999999994h-3.3599999999999994z m0 6.639999999999999v-3.3599999999999994h3.3599999999999994v3.3599999999999994h-3.3599999999999994z m0 13.36v-3.3599999999999994h3.3599999999999994v3.3599999999999994h-3.3599999999999994z m0-6.719999999999999v-3.2833333333333314h3.3599999999999994v3.2833333333333314h-3.3599999999999994z m26.64 6.719999999999999v-3.3599999999999994h3.3599999999999994v3.3599999999999994h-3.3599999999999994z m0 3.280000000000001h3.3599999999999994v3.3599999999999994h-3.3599999999999994v-3.3599999999999994z' })
                )
            );
        }
    }]);

    return MdRoundedCorner;
}(React.Component);

exports.default = MdRoundedCorner;
module.exports = exports['default'];