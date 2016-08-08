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

var FaForumbee = function (_React$Component) {
    _inherits(FaForumbee, _React$Component);

    function FaForumbee() {
        _classCallCheck(this, FaForumbee);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaForumbee).apply(this, arguments));
    }

    _createClass(FaForumbee, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm23.705714285714286 3.3485714285714283q-7.0771428571428565 2.7000000000000006-12.41 8.09142857142857t-7.991428571428571 12.511428571428569q-0.44714285714285795-1.985714285714284-0.44714285714285795-3.9285714285714235 0-4.642857142857142 2.2885714285714287-8.582857142857144t6.217142857142859-6.2285714285714295 8.57142857142857-2.285714285714286q1.8285714285714292 0 3.7714285714285722 0.42285714285714304z m6.0042857142857144 2.657142857142857q2.0757142857142874 1.4500000000000002 3.6599999999999966 3.4571428571428573-8.685714285714287 2.524285714285714-15.057142857142857 8.942857142857143t-8.850000000000001 15.100000000000005q-2.0757142857142856-1.6085714285714303-3.460000000000001-3.617142857142859 2.5-8.614285714285714 8.817142857142857-14.977142857142857t14.88857142857143-8.905714285714286z m-16.36142857142857 29.774285714285718q2.567142857142855-7.9471428571428575 8.471428571428568-13.885714285714286t13.82857142857143-8.571428571428571q0.8928571428571459 2.0542857142857134 1.2057142857142864 4.354285714285714-6.517142857142858 2.678571428571427-11.517142857142858 7.699999999999999t-7.657142857142858 11.564285714285717q-2.297142857142857-0.3142857142857167-4.328571428571429-1.1599999999999966z m23.794285714285714 1.2957142857142827q-4.308571428571426-1.1142857142857139-8.191428571428574-2.567142857142855-3.0142857142857125 1.875714285714288-6.4714285714285715 2.3885714285714315 2.4314285714285724-4.575714285714284 6.114285714285714-8.271428571428572t8.23714285714286-6.1485714285714295q-0.46857142857142975 3.3928571428571423-2.2542857142857144 6.34 1.451428571428572 3.9057142857142857 2.567142857142855 8.257142857142856z' })
                )
            );
        }
    }]);

    return FaForumbee;
}(React.Component);

exports.default = FaForumbee;
module.exports = exports['default'];