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

var MdPhotoSizeSelectActual = function (_React$Component) {
    _inherits(MdPhotoSizeSelectActual, _React$Component);

    function MdPhotoSizeSelectActual() {
        _classCallCheck(this, MdPhotoSizeSelectActual);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdPhotoSizeSelectActual).apply(this, arguments));
    }

    _createClass(MdPhotoSizeSelectActual, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm8.360000000000001 28.36h23.28333333333334l-7.5-10-5.783333333333335 7.5-4.216666666666667-5z m26.64-23.36q1.25 0 2.3049999999999997 1.0549999999999997t1.0549999999999997 2.3049999999999997v23.283333333333335q0 1.2499999999999964-1.0549999999999997 2.303333333333331t-2.3049999999999997 1.0533333333333346h-30q-1.3283333333333331 0-2.3433333333333333-1.0133333333333354t-1.0166666666666666-2.3433333333333337v-23.28333333333333q0-1.2499999999999982 1.0566666666666666-2.303333333333331t2.3033333333333332-1.0566666666666684h30z' })
                )
            );
        }
    }]);

    return MdPhotoSizeSelectActual;
}(React.Component);

exports.default = MdPhotoSizeSelectActual;
module.exports = exports['default'];