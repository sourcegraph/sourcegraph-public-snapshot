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

var MdDonutSmall = function (_React$Component) {
    _inherits(MdDonutSmall, _React$Component);

    function MdDonutSmall() {
        _classCallCheck(this, MdDonutSmall);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdDonutSmall).apply(this, arguments));
    }

    _createClass(MdDonutSmall, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm21.64 24.766666666666666q2.1099999999999994-0.6266666666666652 3.125-3.126666666666665h11.873333333333335q-0.625 6.016666666666666-4.726666666666667 10.233333333333334t-10.273333333333333 4.766666666666666v-11.873333333333335z m3.125-6.406666666666666q-0.9400000000000013-2.5-3.126666666666665-3.125v-11.873333333333333q6.171666666666667 0.5466666666666673 10.273333333333333 4.7666666666666675t4.726666666666667 10.23333333333333h-11.87166666666667z m-6.406666666666666-3.126666666666665q-1.3283333333333331 0.5500000000000007-2.3433333333333337 1.8766666666666652t-1.0150000000000006 2.8900000000000006 1.0166666666666657 2.8900000000000006 2.3433333333333337 1.875v11.873333333333335q-6.326666666666666-0.6216666666666697-10.663333333333332-5.388333333333335t-4.336666666666667-11.25 4.336666666666667-11.25 10.663333333333332-5.390000000000001v11.873333333333335z' })
                )
            );
        }
    }]);

    return MdDonutSmall;
}(React.Component);

exports.default = MdDonutSmall;
module.exports = exports['default'];