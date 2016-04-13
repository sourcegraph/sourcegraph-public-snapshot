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

var FaFolder = function (_React$Component) {
    _inherits(FaFolder, _React$Component);

    function FaFolder() {
        _classCallCheck(this, FaFolder);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaFolder).apply(this, arguments));
    }

    _createClass(FaFolder, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm38.57142857142857 13.571428571428571v15.714285714285714q0 2.0528571428571425-1.471428571428575 3.528571428571432t-3.528571428571425 1.471428571428568h-27.142857142857142q-2.0528571428571425 0-3.528571428571429-1.471428571428568t-1.4714285714285695-3.528571428571432v-21.42857142857143q0-2.0528571428571425 1.4714285714285718-3.5285714285714285t3.5285714285714285-1.4714285714285684h7.142857142857144q2.0528571428571425 0 3.5285714285714285 1.471428571428572t1.4714285714285715 3.5285714285714285v0.7142857142857135h14.999999999999996q2.0528571428571425 0 3.528571428571432 1.4714285714285715t1.471428571428568 3.5285714285714285z' })
                )
            );
        }
    }]);

    return FaFolder;
}(React.Component);

exports.default = FaFolder;
module.exports = exports['default'];