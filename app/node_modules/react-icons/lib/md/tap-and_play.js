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

var MdTapAndPlay = function (_React$Component) {
    _inherits(MdTapAndPlay, _React$Component);

    function MdTapAndPlay() {
        _classCallCheck(this, MdTapAndPlay);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdTapAndPlay).apply(this, arguments));
    }

    _createClass(MdTapAndPlay, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm28.36 1.7166666666666668q1.3283333333333331 0 2.3049999999999997 0.9783333333333335t0.975000000000005 2.3049999999999997v28.36q0 1.3283333333333331-0.9766666666666666 2.3049999999999997t-2.3066666666666684 0.9750000000000014h-3.5166666666666657q-0.23333333333333428-3.3599999999999994-1.5616666666666674-6.640000000000001h5.076666666666668v-21.636666666666667h-16.711666666666673v10q-1.9549999999999983-0.8599999999999994-3.2833333333333314-1.0933333333333337v-12.264999999999999q0-1.3283333333333331 0.9766666666666666-2.3433333333333333t2.3049999999999997-1.0166666666666666z m-25 18.28333333333333q7.578333333333333 0 12.93 5.350000000000001t5.350000000000001 13.009999999999998h-3.280000000000001q0-6.171666666666667-4.413333333333334-10.586666666666666t-10.586666666666668-4.413333333333334v-3.3599999999999994z m8.881784197001252e-16 13.36q2.033333333333333 0 3.5166666666666666 1.4833333333333343t1.4833333333333325 3.5166666666666657h-5v-5z m0-6.719999999999999q4.843333333333333 0 8.241666666666667 3.4383333333333326t3.3983333333333334 8.283333333333331h-3.3599999999999994q0-3.4399999999999977-2.421666666666667-5.899999999999999t-5.858333333333333-2.461666666666666v-3.3599999999999994z' })
                )
            );
        }
    }]);

    return MdTapAndPlay;
}(React.Component);

exports.default = MdTapAndPlay;
module.exports = exports['default'];