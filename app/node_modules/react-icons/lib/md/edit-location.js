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

var MdEditLocation = function (_React$Component) {
    _inherits(MdEditLocation, _React$Component);

    function MdEditLocation() {
        _classCallCheck(this, MdEditLocation);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdEditLocation).apply(this, arguments));
    }

    _createClass(MdEditLocation, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm24.843333333333334 12.578333333333333q0.466666666666665-0.39000000000000057 0-0.8599999999999994l-1.5633333333333326-1.5616666666666674q-0.466666666666665-0.4666666666666668-0.8599999999999994 0l-1.1716666666666669 1.1716666666666669 2.418333333333333 2.421666666666667z m-7.421666666666667 7.421666666666667l5.546666666666667-5.546666666666667-2.418333333333333-2.42-5.550000000000001 5.543333333333335v2.423333333333332h2.421666666666667z m2.578333333333333-16.64q4.843333333333334 8.881784197001252e-16 8.241666666666667 3.4000000000000012t3.3999999999999986 8.239999999999998q0 2.421666666666667-1.2133333333333347 5.546666666666667t-2.928333333333331 5.859999999999999-3.3999999999999986 5.116666666666667-2.8500000000000014 3.788333333333334l-1.25 1.3283333333333331q-0.46999999999999886-0.5466666666666669-1.25-1.4450000000000003t-2.8166666666666647-3.594999999999999-3.5500000000000025-5.233333333333334-2.7733333333333334-5.741666666666667-1.253333333333332-5.625q0-4.843333333333334 3.4000000000000004-8.241666666666667t8.240000000000002-3.4000000000000004z' })
                )
            );
        }
    }]);

    return MdEditLocation;
}(React.Component);

exports.default = MdEditLocation;
module.exports = exports['default'];