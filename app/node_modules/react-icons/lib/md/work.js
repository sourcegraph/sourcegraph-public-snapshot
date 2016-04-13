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

var MdWork = function (_React$Component) {
    _inherits(MdWork, _React$Component);

    function MdWork() {
        _classCallCheck(this, MdWork);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdWork).apply(this, arguments));
    }

    _createClass(MdWork, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm23.36 10v-3.3599999999999994h-6.716666666666669v3.3599999999999994h6.716666666666669z m10 0q1.4066666666666663 0 2.3433333333333337 0.9766666666666666t0.9383333333333326 2.383333333333333v18.283333333333335q0 1.4049999999999976-0.9383333333333326 2.3833333333333364t-2.3433333333333337 0.9733333333333292h-26.716666666666665q-1.408333333333334 0-2.3450000000000006-0.9750000000000014t-0.94-2.383333333333333v-18.28333333333333q0-1.4049999999999994 0.938333333333333-2.383333333333333t2.3433333333333337-0.9750000000000014h6.716666666666669v-3.3599999999999994q0-1.4066666666666663 0.9399999999999995-2.3433333333333337t2.341666666666667-0.9383333333333335h6.716666666666669q1.408333333333335 0 2.344999999999999 0.9383333333333335t0.9383333333333326 2.3433333333333337v3.3599999999999994h6.716666666666669z' })
                )
            );
        }
    }]);

    return MdWork;
}(React.Component);

exports.default = MdWork;
module.exports = exports['default'];