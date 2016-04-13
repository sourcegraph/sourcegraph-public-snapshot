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

var FaPencil = function (_React$Component) {
    _inherits(FaPencil, _React$Component);

    function FaPencil() {
        _classCallCheck(this, FaPencil);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaPencil).apply(this, arguments));
    }

    _createClass(FaPencil, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm10.959999999999999 34.285714285714285l2.031428571428572-2.0314285714285703-5.245714285714286-5.2457142857142856-2.0314285714285703 2.0314285714285703v2.3885714285714315h2.8571428571428568v2.857142857142854h2.388571428571428z m11.674285714285718-20.714285714285715q0-0.4914285714285711-0.4914285714285711-0.4914285714285711-0.2228571428571442 0-0.379999999999999 0.15714285714285658l-12.100000000000001 12.097142857142858q-0.1542857142857148 0.15714285714285836-0.1542857142857148 0.379999999999999 0 0.4914285714285711 0.4914285714285711 0.4914285714285711 0.22285714285714242 0 0.3800000000000008-0.15714285714285836l12.100000000000001-12.097142857142858q0.15428571428571303-0.15714285714285658 0.15428571428571303-0.3800000000000008z m-1.2057142857142864-4.285714285714283l9.285714285714285 9.285714285714286-18.571428571428573 18.571428571428573h-9.285714285714285v-9.285714285714285z m15.245714285714286 2.1428571428571423q0 1.1828571428571433-0.8257142857142838 2.008571428571429l-3.7057142857142864 3.7057142857142846-9.285714285714288-9.285714285714285 3.7057142857142864-3.6828571428571433q0.8028571428571425-0.8485714285714288 2.008571428571429-0.8485714285714288 1.1828571428571415 0 2.0314285714285703 0.8485714285714288l5.245714285714289 5.222857142857142q0.8257142857142838 0.8714285714285719 0.8257142857142838 2.031428571428572z' })
                )
            );
        }
    }]);

    return FaPencil;
}(React.Component);

exports.default = FaPencil;
module.exports = exports['default'];