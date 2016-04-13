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

var MdPhotoFilter = function (_React$Component) {
    _inherits(MdPhotoFilter, _React$Component);

    function MdPhotoFilter() {
        _classCallCheck(this, MdPhotoFilter);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdPhotoFilter).apply(this, arguments));
    }

    _createClass(MdPhotoFilter, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm31.71666666666667 16.64h3.2833333333333314v15q0 1.3283333333333331-0.9766666666666666 2.3433333333333337t-2.306666666666665 1.0166666666666657h-23.35666666666667q-1.3283333333333314 0-2.343333333333332-1.0166666666666657t-1.0166666666666675-2.34333333333333v-23.28333333333334q0-1.3266666666666653 1.0166666666666666-2.341666666666665t2.3400000000000007-1.0150000000000006h15.000000000000002v3.3599999999999994h-15v23.283333333333335h23.36v-15z m-11.091666666666669-5.780000000000001l2.0333333333333314 4.375 4.295000000000002 1.9549999999999983-4.296666666666667 1.9533333333333331-2.0333333333333314 4.375-1.9499999999999993-4.375-4.376666666666667-1.9533333333333331 4.374999999999998-1.9533333333333331z m7.890000000000001 3.9833333333333343l-0.9400000000000013-2.1883333333333326-2.1883333333333326-1.0166666666666657 2.1883333333333326-0.9366666666666674 0.9383333333333326-2.1866666666666674 1.0133333333333319 2.1883333333333326 2.1883333333333326 0.9399999999999995-2.1866666666666674 1.0166666666666675z' })
                )
            );
        }
    }]);

    return MdPhotoFilter;
}(React.Component);

exports.default = MdPhotoFilter;
module.exports = exports['default'];