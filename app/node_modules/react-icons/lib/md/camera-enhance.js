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

var MdCameraEnhance = function (_React$Component) {
    _inherits(MdCameraEnhance, _React$Component);

    function MdCameraEnhance() {
        _classCallCheck(this, MdCameraEnhance);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdCameraEnhance).apply(this, arguments));
    }

    _createClass(MdCameraEnhance, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm20 28.36l-2.1099999999999994-4.609999999999999-4.533333333333335-2.1099999999999994 4.533333333333335-2.0333333333333314 2.1099999999999994-4.606666666666669 2.1099999999999994 4.608333333333334 4.533333333333335 2.0333333333333314-4.533333333333335 2.1083333333333343z m0 1.6400000000000006q3.4383333333333326 0 5.899999999999999-2.461666666666666t2.460000000000001-5.899999999999999-2.460000000000001-5.855-5.899999999999999-2.4250000000000007-5.899999999999999 2.421666666666667-2.460000000000001 5.859999999999999 2.461666666666666 5.899999999999999 5.898333333333333 2.460000000000001z m-5-25h10l3.046666666666667 3.3599999999999994h5.313333333333333q1.3283333333333331 0 2.3049999999999997 0.9766666666666666t0.9750000000000014 2.3049999999999997v20q0 1.3283333333333331-0.9766666666666666 2.3433333333333337t-2.306666666666665 1.0166666666666657h-26.713333333333335q-1.330000000000001 0-2.3066666666666675-1.0166666666666657t-0.9766666666666666-2.3416666666666686v-20q0-1.3283333333333331 0.9766666666666666-2.3049999999999997t2.3050000000000006-0.9766666666666666h5.313333333333334z' })
                )
            );
        }
    }]);

    return MdCameraEnhance;
}(React.Component);

exports.default = MdCameraEnhance;
module.exports = exports['default'];