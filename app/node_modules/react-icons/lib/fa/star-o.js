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

var FaStarO = function (_React$Component) {
    _inherits(FaStarO, _React$Component);

    function FaStarO() {
        _classCallCheck(this, FaStarO);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaStarO).apply(this, arguments));
    }

    _createClass(FaStarO, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm26.808571428571426 22.41l6.828571428571429-6.628571428571428-9.418571428571429-1.3857142857142861-4.218571428571426-8.528571428571428-4.21857142857143 8.528571428571428-9.42 1.3857142857142861 6.828571428571429 6.628571428571428-1.6285714285714281 9.397142857142857 8.438571428571429-4.4428571428571395 8.414285714285715 4.442857142857143z m11.762857142857143-7.967142857142857q0 0.4900000000000002-0.5799999999999983 1.0714285714285712l-8.102857142857143 7.9 1.9200000000000017 11.16q0.022857142857141355 0.15714285714285836 0.022857142857141355 0.4471428571428575 0 1.1142857142857139-0.9142857142857146 1.1142857142857139-0.42571428571428527 0-0.894285714285715-0.2657142857142887l-10.022857142857141-5.269999999999996-10.022857142857143 5.271428571428572q-0.4914285714285711 0.2657142857142887-0.8928571428571423 0.2657142857142887-0.47142857142857153 0-0.7042857142857137-0.32428571428571473t-0.23428571428571487-0.7942857142857136q0-0.134285714285717 0.042857142857142705-0.4471428571428575l1.9214285714285708-11.157142857142858-8.124285714285714-7.904285714285718q-0.5571428571428569-0.6028571428571432-0.5571428571428569-1.0714285714285712 0-0.8257142857142856 1.2485714285714287-1.0285714285714285l11.205714285714286-1.6285714285714281 5.022857142857145-10.157142857142858q0.42285714285713993-0.9099999999999994 1.0942857142857108-0.9099999999999994t1.0942857142857143 0.9142857142857143l5.022857142857145 10.157142857142858 11.205714285714286 1.6285714285714281q1.25 0.20285714285714285 1.25 1.0285714285714285z' })
                )
            );
        }
    }]);

    return FaStarO;
}(React.Component);

exports.default = FaStarO;
module.exports = exports['default'];