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

var TiDeviceDesktop = function (_React$Component) {
    _inherits(TiDeviceDesktop, _React$Component);

    function TiDeviceDesktop() {
        _classCallCheck(this, TiDeviceDesktop);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiDeviceDesktop).apply(this, arguments));
    }

    _createClass(TiDeviceDesktop, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm35 1.6666666666666667h-30c-2.7566666666666664 0-5 2.243333333333333-5 5v18.333333333333336c0 2.7566666666666677 2.2433333333333336 5 5 5h10v3.333333333333332h-5c-0.9199999999999999 0-1.666666666666666 0.7466666666666697-1.666666666666666 1.6666666666666643s0.7466666666666661 1.6666666666666643 1.666666666666666 1.6666666666666643h20c0.9200000000000017 0 1.6666666666666679-0.7466666666666697 1.6666666666666679-1.6666666666666643s-0.7466666666666661-1.6666666666666643-1.6666666666666679-1.6666666666666643h-5v-3.333333333333332h10c2.7566666666666677 0 5-2.2433333333333323 5-5v-18.333333333333336c0-2.7566666666666677-2.2433333333333323-5.000000000000001-5-5.000000000000001z m-11.666666666666668 31.666666666666668h-6.666666666666668v-3.333333333333332h6.666666666666668v3.333333333333332z m13.333333333333332-8.333333333333336c0 0.9166666666666679-0.75 1.6666666666666679-1.6666666666666643 1.6666666666666679h-30c-0.916666666666667 0-1.666666666666667-0.75-1.666666666666667-1.6666666666666679v-18.333333333333332c0-0.9166666666666679 0.75-1.6666666666666679 1.666666666666667-1.6666666666666679h30c0.9166666666666643 0 1.6666666666666643 0.75 1.6666666666666643 1.666666666666667v18.333333333333336z m-3.3333333333333286-18.333333333333332h-26.666666666666668c-0.9166666666666679-8.881784197001252e-16-1.6666666666666679 0.7499999999999991-1.6666666666666679 1.666666666666666v13.333333333333334c0 0.9166666666666679 0.75 1.6666666666666679 1.666666666666667 1.6666666666666679h26.666666666666668c0.9166666666666643 0 1.6666666666666643-0.75 1.6666666666666643-1.6666666666666679v-13.333333333333334c0-0.916666666666667-0.75-1.666666666666667-1.6666666666666643-1.666666666666667z m0 15h-26.666666666666668v-13.333333333333334h26.666666666666668v13.333333333333334z' })
                )
            );
        }
    }]);

    return TiDeviceDesktop;
}(React.Component);

exports.default = TiDeviceDesktop;
module.exports = exports['default'];