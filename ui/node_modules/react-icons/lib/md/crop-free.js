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

var MdCropFree = function (_React$Component) {
    _inherits(MdCropFree, _React$Component);

    function MdCropFree() {
        _classCallCheck(this, MdCropFree);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdCropFree).apply(this, arguments));
    }

    _createClass(MdCropFree, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm31.640000000000004 5q1.3283333333333367 0 2.34333333333333 1.0166666666666666t1.0166666666666657 2.3400000000000007v6.643333333333333h-3.361666666666668v-6.643333333333333h-6.638333333333332v-3.3566666666666674h6.638333333333335z m0 26.64v-6.640000000000001h3.359999999999996v6.640000000000001q0 1.3283333333333331-1.0166666666666657 2.3433333333333337t-2.3416666666666686 1.0166666666666657h-6.641666666666666v-3.361666666666668h6.641666666666666z m-23.28-6.640000000000001v6.640000000000001h6.639999999999997v3.3599999999999994h-6.639999999999999q-1.3283333333333331 0-2.3433333333333337-1.0166666666666657t-1.0166666666666675-2.34v-6.643333333333334h3.3616666666666664z m-3.360000000000003-16.64q0-1.3283333333333314 1.0166666666666666-2.343333333333332t2.3400000000000007-1.0166666666666675h6.643333333333333v3.3616666666666664h-6.643333333333333v6.638333333333334h-3.3566666666666674v-6.636666666666667z' })
                )
            );
        }
    }]);

    return MdCropFree;
}(React.Component);

exports.default = MdCropFree;
module.exports = exports['default'];