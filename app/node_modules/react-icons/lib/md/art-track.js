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

var MdArtTrack = function (_React$Component) {
    _inherits(MdArtTrack, _React$Component);

    function MdArtTrack() {
        _classCallCheck(this, MdArtTrack);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdArtTrack).apply(this, arguments));
    }

    _createClass(MdArtTrack, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm17.5 25l-3.75-5-2.8900000000000006 3.75-2.1099999999999994-2.5-2.8899999999999997 3.75h11.64z m2.5-10v10q0 1.3283333333333331-1.0166666666666657 2.3433333333333337t-2.3416666666666686 1.0166666666666657h-10q-1.3283333333333331 0-2.3049999999999997-1.0166666666666657t-0.9766666666666657-2.3433333333333337v-10q0-1.3283333333333331 0.9766666666666666-2.3433333333333337t2.3050000000000006-1.0166666666666657h9.999999999999998q1.3283333333333331 0 2.3433333333333337 1.0166666666666657t1.0150000000000006 2.3433333333333337z m3.3599999999999994 13.36v-3.3599999999999994h13.283333333333331v3.3599999999999994h-13.283333333333331z m13.280000000000001-16.720000000000002v3.360000000000003h-13.283333333333331v-3.3599999999999994h13.283333333333331z m0 10h-13.283333333333331v-3.283333333333335h13.283333333333331v3.2833333333333314z' })
                )
            );
        }
    }]);

    return MdArtTrack;
}(React.Component);

exports.default = MdArtTrack;
module.exports = exports['default'];