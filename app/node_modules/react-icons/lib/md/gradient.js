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

var MdGradient = function (_React$Component) {
    _inherits(MdGradient, _React$Component);

    function MdGradient() {
        _classCallCheck(this, MdGradient);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdGradient).apply(this, arguments));
    }

    _createClass(MdGradient, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm31.640000000000004 18.36v-10h-23.28333333333334v10h3.283333333333335v3.2833333333333314h3.3599999999999994v3.356666666666669h3.3599999999999994v-3.3583333333333343h3.2833333333333314v3.3583333333333343h3.356666666666669v-3.3583333333333343h3.361666666666668v-3.2833333333333314h3.283333333333335z m-3.2800000000000047 11.64v-3.3599999999999994h-3.3599999999999994v3.3599999999999994h3.3599999999999994z m-6.719999999999999 0v-3.3599999999999994h-3.2833333333333314v3.3599999999999994h3.2833333333333314z m-6.640000000000001 0v-3.3599999999999994h-3.3599999999999994v3.3599999999999994h3.3599999999999994z m16.64-25q1.3283333333333331 0 2.3433333333333337 1.0166666666666666t1.0166666666666657 2.3400000000000007v23.28333333333333q0 1.326666666666668-1.0166666666666657 2.3416666666666686t-2.3433333333333337 1.0166666666666657h-23.28333333333333q-1.3266666666666689 0-2.3416666666666686-1.0166666666666657t-1.0150000000000006-2.341666666666665v-23.28333333333334q0-1.3266666666666653 1.0166666666666666-2.341666666666665t2.3400000000000007-1.0150000000000006h23.28333333333333z m-20 10h3.3599999999999994v3.3599999999999994h-3.3599999999999994v-3.3599999999999994z m13.36 0h3.3599999999999994v3.3599999999999994h-3.3599999999999994v-3.3599999999999994z m-6.640000000000001 0h3.2833333333333314v3.3599999999999994h3.356666666666669v3.2833333333333314h-3.3583333333333343v-3.2833333333333314h-3.2833333333333314v3.2833333333333314h-3.3583333333333343v-3.2833333333333314h3.3599999999999994v-3.3599999999999994z m10 6.640000000000001v3.3599999999999994h3.2833333333333314v-3.3599999999999994h-3.2833333333333314z m-16.720000000000002 0h-3.283333333333333v3.3599999999999994h3.283333333333333v-3.3599999999999994z' })
                )
            );
        }
    }]);

    return MdGradient;
}(React.Component);

exports.default = MdGradient;
module.exports = exports['default'];