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

var FaDatabase = function (_React$Component) {
    _inherits(FaDatabase, _React$Component);

    function FaDatabase() {
        _classCallCheck(this, FaDatabase);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaDatabase).apply(this, arguments));
    }

    _createClass(FaDatabase, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm20 17.142857142857142q5.289999999999999 0 9.888571428571428-0.9600000000000009t7.254285714285718-2.8342857142857127v3.7942857142857136q0 1.5399999999999991-2.299999999999997 2.8571428571428577t-6.248571428571427 2.0857142857142854-8.594285714285721 0.7714285714285722-8.594285714285714-0.7714285714285722-6.248571428571428-2.0857142857142854-2.3000000000000003-2.8571428571428577v-3.7942857142857136q2.657142857142857 1.8757142857142863 7.2542857142857144 2.8342857142857163t9.888571428571428 0.9599999999999973z m0 17.142857142857142q5.289999999999999 0 9.888571428571428-0.9600000000000009t7.254285714285718-2.834285714285709v3.79428571428571q0 1.5399999999999991-2.299999999999997 2.857142857142854t-6.248571428571427 2.085714285714289-8.594285714285721 0.7714285714285722-8.594285714285714-0.7714285714285722-6.248571428571428-2.085714285714282-2.3000000000000003-2.857142857142861v-3.7942857142857136q2.657142857142857 1.875714285714288 7.2542857142857144 2.8342857142857127t9.888571428571428 0.9600000000000009z m0-8.57142857142857q5.289999999999999 0 9.888571428571428-0.9600000000000009t7.254285714285718-2.8342857142857163v3.794285714285717q0 1.5399999999999991-2.299999999999997 2.8571428571428577t-6.248571428571427 2.0857142857142854-8.594285714285721 0.7714285714285722-8.594285714285714-0.7714285714285722-6.248571428571428-2.0857142857142854-2.3000000000000003-2.8571428571428577v-3.7942857142857136q2.657142857142857 1.8742857142857154 7.2542857142857144 2.8342857142857127t9.888571428571428 0.9600000000000009z m0-25.714285714285715q4.642857142857142 0 8.594285714285714 0.7714285714285715t6.248571428571427 2.085714285714286 2.3000000000000043 2.857142857142857v2.8571428571428568q0 1.540000000000001-2.299999999999997 2.8571428571428577t-6.248571428571427 2.0857142857142854-8.594285714285721 0.7714285714285722-8.594285714285714-0.7714285714285722-6.248571428571428-2.0857142857142854-2.3000000000000003-2.8571428571428577v-2.8571428571428568q0-1.54 2.3000000000000003-2.857142857142857t6.247142857142857-2.085714285714286 8.595714285714285-0.7714285714285714z' })
                )
            );
        }
    }]);

    return FaDatabase;
}(React.Component);

exports.default = FaDatabase;
module.exports = exports['default'];