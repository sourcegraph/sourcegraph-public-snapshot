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

var MdGpsOff = function (_React$Component) {
    _inherits(MdGpsOff, _React$Component);

    function MdGpsOff() {
        _classCallCheck(this, MdGpsOff);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdGpsOff).apply(this, arguments));
    }

    _createClass(MdGpsOff, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm27.11 29.21666666666667l-16.326666666666668-16.325000000000003q-2.424999999999999 3.201666666666668-2.424999999999999 7.108333333333334 0 4.843333333333334 3.4000000000000004 8.241666666666667t8.241666666666667 3.3999999999999986q3.9066666666666663 0 7.109999999999999-2.423333333333332z m-22.11-22.105l2.1100000000000003-2.111666666666668 27.89 27.890000000000008-2.1099999999999923 2.1099999999999923-3.4383333333333326-3.4383333333333326q-3.5133333333333354 2.8900000000000006-7.811666666666667 3.3599999999999994v3.4383333333333326h-3.2833333333333314v-3.4383333333333326q-5.233333333333334-0.5466666666666669-8.983333333333333-4.296666666666667t-4.296666666666667-8.983333333333334h-3.44000000000001v-3.2833333333333314h3.4383333333333344q0.4666666666666668-4.296666666666667 3.360000000000001-7.813333333333333z m29.921666666666667 11.25h3.4383333333333326v3.283333333333335h-3.4383333333333326q-0.39000000000000057 3.0450000000000017-1.6400000000000006 5.311666666666667l-2.5-2.5q0.8583333333333343-2.1099999999999994 0.8583333333333343-4.453333333333333 0-4.841666666666667-3.3999999999999986-8.24t-8.240000000000002-3.4000000000000004q-2.3416666666666686 0-4.449999999999999 0.8616666666666664l-2.5-2.5q2.42-1.25 5.311666666666667-1.6400000000000006v-3.443333333333336h3.283333333333335v3.4416666666666664q5.233333333333334 0.5449999999999999 8.983333333333334 4.295000000000001t4.296666666666667 8.98333333333333z' })
                )
            );
        }
    }]);

    return MdGpsOff;
}(React.Component);

exports.default = MdGpsOff;
module.exports = exports['default'];