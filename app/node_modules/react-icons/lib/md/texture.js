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

var MdTexture = function (_React$Component) {
    _inherits(MdTexture, _React$Component);

    function MdTexture() {
        _classCallCheck(this, MdTexture);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdTexture).apply(this, arguments));
    }

    _createClass(MdTexture, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm15.466666666666667 35l19.53333333333333-19.53333333333333v4.766666666666666l-14.766666666666666 14.766666666666666h-4.7616666666666685z m19.53333333333333-3.359999999999996q0 1.3283333333333367-1.0166666666666657 2.34333333333333t-2.3416666666666686 1.0166666666666657h-3.2833333333333314l6.641666666666666-6.643333333333334v3.283333333333335z m-26.64-26.640000000000004h3.283333333333335l-6.643333333333334 6.640000000000001v-3.283333333333333q0-1.326666666666667 1.0166666666666666-2.341666666666667t2.3433333333333346-1.0150000000000006z m11.405000000000001 0h4.763333333333335l-19.53166666666667 19.533333333333335v-4.766666666666666z m12.733333333333334 0.1566666666666663q1.9533333333333331 0.5466666666666669 2.421666666666667 2.3433333333333337l-27.42 27.343333333333334q-1.7966666666666669-0.5466666666666669-2.3433333333333337-2.3433333333333337z' })
                )
            );
        }
    }]);

    return MdTexture;
}(React.Component);

exports.default = MdTexture;
module.exports = exports['default'];