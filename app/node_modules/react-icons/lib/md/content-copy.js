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

var MdContentCopy = function (_React$Component) {
    _inherits(MdContentCopy, _React$Component);

    function MdContentCopy() {
        _classCallCheck(this, MdContentCopy);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdContentCopy).apply(this, arguments));
    }

    _createClass(MdContentCopy, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm31.640000000000004 35v-23.36h-18.28333333333334v23.36h18.283333333333335z m0-26.64q1.3283333333333367 0 2.34333333333333 0.9766666666666666t1.0166666666666657 2.3050000000000015v23.358333333333334q0 1.3299999999999983-1.0166666666666657 2.344999999999999t-2.3433333333333337 1.0166666666666657h-18.283333333333335q-1.3266666666666662 0-2.341666666666667-1.0166666666666657t-1.0149999999999988-2.344999999999999v-23.35666666666667q0-1.3283333333333314 1.0166666666666657-2.304999999999998t2.341666666666667-0.9766666666666666h18.28333333333333z m-5-6.720000000000001v3.360000000000001h-20v23.36h-3.283333333333333v-23.36q0-1.3283333333333331 0.9783333333333335-2.3433333333333333t2.3066666666666666-1.0166666666666666h20z' })
                )
            );
        }
    }]);

    return MdContentCopy;
}(React.Component);

exports.default = MdContentCopy;
module.exports = exports['default'];