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

var FaPlayCircle = function (_React$Component) {
    _inherits(FaPlayCircle, _React$Component);

    function FaPlayCircle() {
        _classCallCheck(this, FaPlayCircle);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaPlayCircle).apply(this, arguments));
    }

    _createClass(FaPlayCircle, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm20 2.857142857142857q4.665714285714287 0 8.604285714285716 2.3000000000000003t6.238571428571426 6.2371428571428575 2.3000000000000043 8.605714285714285-2.299999999999997 8.602857142857143-6.238571428571429 6.238571428571429-8.60428571428572 2.297142857142859-8.604285714285714-2.299999999999997-6.238571428571428-6.237142857142857-2.3000000000000003-8.601428571428578 2.3000000000000003-8.605714285714287 6.238571428571428-6.237142857142856 8.604285714285714-2.3000000000000003z m8.57142857142857 18.37142857142857q0.7142857142857153-0.4028571428571439 0.7142857142857153-1.2285714285714278t-0.7142857142857153-1.2285714285714278l-12.142857142857142-7.142857142857142q-0.6914285714285722-0.4228571428571435-1.4285714285714288-0.02142857142857224-0.7142857142857117 0.4242857142857144-0.7142857142857117 1.25v14.285714285714288q0 0.8257142857142838 0.7142857142857135 1.25 0.35714285714285765 0.17857142857142705 0.7142857142857135 0.17857142857142705 0.3800000000000008 0 0.7142857142857135-0.1999999999999993z' })
                )
            );
        }
    }]);

    return FaPlayCircle;
}(React.Component);

exports.default = FaPlayCircle;
module.exports = exports['default'];