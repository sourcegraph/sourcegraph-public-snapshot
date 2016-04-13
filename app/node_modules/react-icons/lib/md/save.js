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

var MdSave = function (_React$Component) {
    _inherits(MdSave, _React$Component);

    function MdSave() {
        _classCallCheck(this, MdSave);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdSave).apply(this, arguments));
    }

    _createClass(MdSave, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm25 15v-6.639999999999999h-16.64v6.639999999999999h16.64z m-5 16.64q2.0333333333333314 0 3.5166666666666657-1.4833333333333343t1.4833333333333343-3.5166666666666657-1.4833333333333343-3.5166666666666657-3.5166666666666657-1.4833333333333343-3.5166666666666657 1.4833333333333343-1.4833333333333343 3.5166666666666657 1.4833333333333343 3.5166666666666657 3.5166666666666657 1.4833333333333343z m8.36-26.64l6.640000000000001 6.640000000000001v20q0 1.3283333333333331-1.0166666666666657 2.3433333333333337t-2.3416666666666686 1.0166666666666657h-23.28333333333333q-1.405000000000002 0-2.3833333333333346-1.0166666666666657t-0.9749999999999996-2.34333333333333v-23.28333333333334q0-1.3266666666666653 0.9766666666666666-2.341666666666665t2.3833333333333346-1.0150000000000006h20z' })
                )
            );
        }
    }]);

    return MdSave;
}(React.Component);

exports.default = MdSave;
module.exports = exports['default'];