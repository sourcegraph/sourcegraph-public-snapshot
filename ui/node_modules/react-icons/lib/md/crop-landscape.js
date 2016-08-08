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

var MdCropLandscape = function (_React$Component) {
    _inherits(MdCropLandscape, _React$Component);

    function MdCropLandscape() {
        _classCallCheck(this, MdCropLandscape);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdCropLandscape).apply(this, arguments));
    }

    _createClass(MdCropLandscape, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm31.640000000000004 28.36v-16.71666666666667h-23.28333333333334v16.71666666666667h23.283333333333335z m0-20q1.3283333333333367 0 2.34333333333333 0.9766666666666666t1.0166666666666657 2.3050000000000015v16.71666666666667q0 1.3299999999999983-1.0166666666666657 2.306666666666665t-2.3433333333333337 0.9750000000000014h-23.28333333333333q-1.3266666666666689 0-2.3416666666666686-0.9766666666666666t-1.0150000000000006-2.3049999999999997v-16.715000000000003q0-1.3299999999999983 1.0166666666666666-2.306666666666665t2.3416666666666677-0.9766666666666666h23.283333333333335z' })
                )
            );
        }
    }]);

    return MdCropLandscape;
}(React.Component);

exports.default = MdCropLandscape;
module.exports = exports['default'];