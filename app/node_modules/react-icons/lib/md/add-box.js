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

var MdAddBox = function (_React$Component) {
    _inherits(MdAddBox, _React$Component);

    function MdAddBox() {
        _classCallCheck(this, MdAddBox);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdAddBox).apply(this, arguments));
    }

    _createClass(MdAddBox, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm28.36 21.64v-3.2833333333333314h-6.716666666666669v-6.716666666666669h-3.2833333333333314v6.716666666666669h-6.716666666666669v3.2833333333333314h6.716666666666669v6.716666666666669h3.2833333333333314v-6.716666666666669h6.716666666666669z m3.2800000000000047-16.64q1.3283333333333367 0 2.34333333333333 1.0166666666666666t1.0166666666666657 2.3400000000000007v23.28333333333333q0 1.326666666666668-1.0166666666666657 2.3416666666666686t-2.3433333333333337 1.0166666666666657h-23.28333333333333q-1.405000000000002 0-2.3833333333333346-1.0166666666666657t-0.9733333333333345-2.341666666666665v-23.28333333333334q0-1.3266666666666653 0.9749999999999996-2.341666666666665t2.383333333333333-1.0150000000000006h23.28333333333333z' })
                )
            );
        }
    }]);

    return MdAddBox;
}(React.Component);

exports.default = MdAddBox;
module.exports = exports['default'];