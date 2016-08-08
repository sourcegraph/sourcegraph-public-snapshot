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

var TiDeviceTablet = function (_React$Component) {
    _inherits(TiDeviceTablet, _React$Component);

    function TiDeviceTablet() {
        _classCallCheck(this, TiDeviceTablet);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiDeviceTablet).apply(this, arguments));
    }

    _createClass(TiDeviceTablet, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm28.333333333333336 6.666666666666667h-15.000000000000002c-0.9166666666666661 0-1.666666666666666 0.75-1.666666666666666 1.666666666666667v20c0 0.9166666666666679 0.75 1.6666666666666679 1.666666666666666 1.6666666666666679h5.833333333333334c0 0.9216666666666669 0.7466666666666661 1.6666666666666679 1.6666666666666679 1.6666666666666679s1.6666666666666679-0.745000000000001 1.6666666666666679-1.6666666666666679h5.833333333333332c0.9166666666666679 0 1.6666666666666679-0.75 1.6666666666666679-1.6666666666666679v-20c0-0.9166666666666687-0.75-1.6666666666666687-1.6666666666666679-1.6666666666666687z m0 21.666666666666668h-15.000000000000002v-20h15.000000000000002v20z m1.6666666666666643-26.666666666666668h-18.333333333333332c-2.7566666666666677-1.1102230246251565e-15-5.000000000000001 2.2433333333333323-5.000000000000001 4.999999999999999v25c0 2.756666666666664 2.243333333333333 5.0000000000000036 5.000000000000001 5.0000000000000036h18.333333333333336c2.756666666666664 0 4.9999999999999964-2.2433333333333323 4.9999999999999964-5v-25.000000000000004c0-2.7566666666666677-2.2433333333333323-5.000000000000001-5-5.000000000000001z m1.6666666666666679 30c0 0.9166666666666679-0.75 1.6666666666666679-1.6666666666666679 1.6666666666666679h-18.333333333333332c-0.9166666666666661 0-1.666666666666666-0.75-1.666666666666666-1.6666666666666679v-25c0-0.9166666666666679 0.75-1.6666666666666679 1.666666666666666-1.6666666666666679h18.333333333333336c0.9166666666666679 0 1.6666666666666679 0.75 1.6666666666666679 1.666666666666667v25z' })
                )
            );
        }
    }]);

    return TiDeviceTablet;
}(React.Component);

exports.default = TiDeviceTablet;
module.exports = exports['default'];