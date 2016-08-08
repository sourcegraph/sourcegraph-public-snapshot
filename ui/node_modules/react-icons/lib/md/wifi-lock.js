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

var MdWifiLock = function (_React$Component) {
    _inherits(MdWifiLock, _React$Component);

    function MdWifiLock() {
        _classCallCheck(this, MdWifiLock);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdWifiLock).apply(this, arguments));
    }

    _createClass(MdWifiLock, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm36.64000000000001 26.64v-2.5q0-1.0166666666666657-0.7416666666666671-1.7583333333333329t-1.759999999999998-0.7433333333333323-1.7583333333333329 0.7416666666666671-0.7433333333333323 1.7566666666666677v2.5h5z m1.7199999999999989 0q0.7033333333333331 0 1.1716666666666669 0.5083333333333329t0.46666666666666856 1.2100000000000009v6.641666666666666q0 0.7033333333333331-0.46666666666666856 1.1716666666666669t-1.173333333333332 0.46666666666666856h-8.358333333333341q-0.7033333333333331 0-1.1716666666666669-0.46666666666666856t-0.466666666666665-1.1716666666666669v-6.640000000000001q0-0.7033333333333331 0.466666666666665-1.211666666666666t1.1716666666666669-0.5100000000000016v-2.5q0-1.716666666666665 1.211666666666666-2.9299999999999997t2.9299999999999997-1.2083333333333321 2.9666666666666686 1.2100000000000009 1.25 2.9299999999999997v2.5z m-4.219999999999999-10.780000000000001q-3.4383333333333326 0-5.859999999999999 2.421666666666667t-2.421666666666667 5.858333333333334v4.766666666666666l-5.858333333333341 7.733333333333341-20-26.640000000000008q8.75-6.638333333333334 20-6.638333333333334t20 6.638333333333334l-4.454999999999998 5.940000000000001q-0.46666666666666856-0.07833333333333314-1.4066666666666663-0.07833333333333314z' })
                )
            );
        }
    }]);

    return MdWifiLock;
}(React.Component);

exports.default = MdWifiLock;
module.exports = exports['default'];