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

var FaNeuter = function (_React$Component) {
    _inherits(FaNeuter, _React$Component);

    function FaNeuter() {
        _classCallCheck(this, FaNeuter);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaNeuter).apply(this, arguments));
    }

    _createClass(FaNeuter, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm32.85714285714286 12.857142857142858q0 4.9328571428571415-3.2928571428571445 8.582857142857144t-8.135714285714286 4.185714285714283v13.66q0 0.3142857142857167-0.1999999999999993 0.5142857142857125t-0.514285714285716 0.20000000000000284h-1.428571428571427q-0.31428571428571317 0-0.514285714285716-0.20000000000000284t-0.1999999999999993-0.5142857142857125v-13.66q-4.842857142857143-0.5357142857142847-8.135714285714286-4.185714285714287t-3.2928571428571436-8.58285714285714q0-2.611428571428572 1.0142857142857151-4.988571428571428t2.747142857142858-4.107142857142858 4.107142857142854-2.7471428571428573 4.988571428571429-1.0142857142857142 4.988571428571429 1.0142857142857142 4.107142857142858 2.7471428571428573 2.7457142857142856 4.107142857142858 1.0157142857142887 4.988571428571428z m-12.857142857142858 10q4.12857142857143 0 7.064285714285713-2.935714285714287t2.9357142857142833-7.064285714285713-2.935714285714287-7.064285714285715-7.064285714285713-2.9357142857142855-7.064285714285715 2.9357142857142855-2.935714285714285 7.064285714285715 2.935714285714287 7.064285714285713 7.064285714285713 2.935714285714287z' })
                )
            );
        }
    }]);

    return FaNeuter;
}(React.Component);

exports.default = FaNeuter;
module.exports = exports['default'];