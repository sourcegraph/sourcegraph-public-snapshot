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

var MdHearing = function (_React$Component) {
    _inherits(MdHearing, _React$Component);

    function MdHearing() {
        _classCallCheck(this, MdHearing);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdHearing).apply(this, arguments));
    }

    _createClass(MdHearing, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm19.14 15q0-1.7166666666666668 1.25-2.9299999999999997t2.9666666666666686-1.211666666666666 2.9333333333333336 1.211666666666666 1.2099999999999973 2.9299999999999997-1.211666666666666 2.9299999999999997-2.9299999999999997 1.211666666666666-2.9666666666666686-1.211666666666666-1.2516666666666652-2.9299999999999997z m-6.406666666666668-10.625q-4.371666666666664 4.375-4.371666666666664 10.625t4.375 10.625l-2.3433333333333355 2.3416666666666686q-5.393333333333333-5.386666666666667-5.393333333333333-12.966666666666669t5.391666666666666-12.966666666666667z m15.626666666666667 28.983333333333334q1.3283333333333331 0 2.3049999999999997-1.0133333333333354t0.975000000000005-2.344999999999999h3.359999999999996q0 2.7366666666666646-1.9500000000000028 4.689999999999998t-4.690000000000001 1.9533333333333331q-1.5633333333333326 0-2.7333333333333343-0.5466666666666669-3.205000000000002-1.6400000000000006-4.611666666666668-5.938333333333333-0.5500000000000007-1.7166666666666686-2.8166666666666664-3.4383333333333326-3.1999999999999993-2.3433333333333337-4.763333333333334-5.233333333333334-1.7949999999999928-3.2066666666666634-1.7949999999999928-6.486666666666665 0-4.921666666666667 3.4000000000000004-8.283333333333333t8.316666666666668-3.358333333333334 8.283333333333331 3.3600000000000008 3.3599999999999994 8.281666666666666h-3.3599999999999994q0-3.5166666666666657-2.383333333333333-5.938333333333333t-5.899999999999999-2.421666666666667-5.936666666666667 2.421666666666667-2.4200000000000017 5.938333333333333q0 2.5 1.326666666666668 4.921666666666667 1.0933333333333337 2.1099999999999994 3.9066666666666663 4.216666666666669 3.125 2.344999999999999 3.9833333333333343 5 1.0166666666666657 2.969999999999999 2.8166666666666664 3.9083333333333314 0.6233333333333348 0.3133333333333326 1.326666666666668 0.3133333333333326z' })
                )
            );
        }
    }]);

    return MdHearing;
}(React.Component);

exports.default = MdHearing;
module.exports = exports['default'];