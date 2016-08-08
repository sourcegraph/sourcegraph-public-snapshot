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

var TiDevicePhone = function (_React$Component) {
    _inherits(TiDevicePhone, _React$Component);

    function TiDevicePhone() {
        _classCallCheck(this, TiDevicePhone);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiDevicePhone).apply(this, arguments));
    }

    _createClass(TiDevicePhone, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm25 5h-11.666666666666666c-2.756666666666666 0-5 2.243333333333334-5 5v20c0 2.7566666666666677 2.243333333333334 5 5 5h11.666666666666666c2.7566666666666677 0 5-2.2433333333333323 5-5v-20c0-2.756666666666667-2.2433333333333323-5-5-5z m1.6666666666666679 25c0 0.9166666666666679-0.75 1.6666666666666679-1.6666666666666679 1.6666666666666679h-11.666666666666666c-0.9166666666666661 0-1.666666666666666-0.75-1.666666666666666-1.6666666666666679v-20c0-0.9166666666666661 0.75-1.666666666666666 1.666666666666666-1.666666666666666h11.666666666666666c0.9166666666666679 0 1.6666666666666679 0.75 1.6666666666666679 1.666666666666666v20z m-3.333333333333332-20h-8.333333333333336c-0.9166666666666661 0-1.666666666666666 0.75-1.666666666666666 1.666666666666666v14.999999999999998c0 0.9166666666666679 0.75 1.6666666666666679 1.666666666666666 1.6666666666666679h2.5c0 0.9216666666666669 0.7466666666666661 1.6666666666666679 1.6666666666666679 1.6666666666666679s1.6666666666666679-0.745000000000001 1.6666666666666679-1.6666666666666679h2.5c0.9166666666666679 0 1.6666666666666679-0.75 1.6666666666666679-1.6666666666666679v-14.999999999999996c0-0.9166666666666661-0.75-1.666666666666666-1.6666666666666679-1.666666666666666z m0 16.666666666666668h-8.333333333333336v-15h8.333333333333336v15z' })
                )
            );
        }
    }]);

    return TiDevicePhone;
}(React.Component);

exports.default = TiDevicePhone;
module.exports = exports['default'];