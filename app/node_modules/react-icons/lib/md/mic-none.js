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

var MdMicNone = function (_React$Component) {
    _inherits(MdMicNone, _React$Component);

    function MdMicNone() {
        _classCallCheck(this, MdMicNone);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdMicNone).apply(this, arguments));
    }

    _createClass(MdMicNone, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm28.828333333333337 18.36h2.8133333333333326q0 4.216666666666669-2.9299999999999997 7.383333333333333t-7.07 3.788333333333334v5.468333333333334h-3.283333333333335v-5.466666666666669q-4.138333333333334-0.6266666666666652-7.066666666666666-3.789999999999999t-2.9333333333333336-7.383333333333333h2.8166666666666664q0 3.671666666666667 2.616666666666667 6.055t6.208333333333332 2.384999999999998 6.213333333333335-2.383333333333333 2.616666666666667-6.056666666666665z m-10.861666666666672-10.156666666666666v10.313333333333333q0 0.783333333333335 0.5883333333333347 1.3666666666666671t1.4450000000000003 0.586666666666666q0.783333333333335 0 1.3666666666666671-0.5466666666666669t0.586666666666666-1.4066666666666663l0.07833333333333314-10.313333333333333q0-0.8600000000000003-0.6266666666666652-1.4450000000000003t-1.4050000000000011-0.586666666666666-1.4066666666666663 0.5866666666666669-0.625 1.4449999999999994z m2.033333333333335 15.156666666666666q-2.0333333333333314 0-3.5166666666666657-1.4833333333333343t-1.4833333333333343-3.518333333333331v-10q0-2.033333333333333 1.4833333333333343-3.5166666666666666t3.5166666666666657-1.4833333333333334 3.5166666666666657 1.4833333333333334 1.4833333333333343 3.5166666666666666v10q0 2.0333333333333314-1.4833333333333343 3.5166666666666657t-3.5166666666666657 1.4833333333333343z' })
                )
            );
        }
    }]);

    return MdMicNone;
}(React.Component);

exports.default = MdMicNone;
module.exports = exports['default'];