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

var FaXing = function (_React$Component) {
    _inherits(FaXing, _React$Component);

    function FaXing() {
        _classCallCheck(this, FaXing);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaXing).apply(this, arguments));
    }

    _createClass(FaXing, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm17.611428571428572 14.88857142857143q-0.2228571428571442 0.40000000000000036-5.737142857142857 10.178571428571429-0.6028571428571432 1.0285714285714285-1.451428571428572 1.0285714285714285h-5.3342857142857145q-0.46857142857142886 0-0.6914285714285713-0.38142857142857167t0-0.8028571428571425l5.647142857142856-10q0.024285714285714022 0 0-0.02285714285714313l-3.5928571428571434-6.2285714285714295q-0.2671428571428569-0.4900000000000002-0.02285714285714313-0.8242857142857147 0.20000000000000018-0.3342857142857145 0.7142857142857144-0.3342857142857145h5.332857142857143q0.8928571428571423 0 1.4714285714285715 1.0057142857142853z m17.991428571428575-14.32857142857143q0.24571428571428555 0.35714285714285676 0 0.8242857142857138l-11.785714285714285 20.847142857142856v0.022857142857141355l7.5 13.725714285714286q0.24571428571428555 0.4471428571428575 0.022857142857141355 0.8257142857142838-0.2228571428571442 0.33428571428571274-0.7142857142857153 0.33428571428571274h-5.334285714285713q-0.937142857142856 0-1.4714285714285715-1.0042857142857144l-7.568571428571428-13.885714285714286q0.3999999999999986-0.7142857142857153 11.852857142857143-21.025714285714287 0.5571428571428569-1.0057142857142858 1.428571428571427-1.0057142857142858h5.380000000000003q0.490000000000002 0 0.6899999999999977 0.3342857142857143z' })
                )
            );
        }
    }]);

    return FaXing;
}(React.Component);

exports.default = FaXing;
module.exports = exports['default'];