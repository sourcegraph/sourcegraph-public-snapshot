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

var TiBatteryHigh = function (_React$Component) {
    _inherits(TiBatteryHigh, _React$Component);

    function TiBatteryHigh() {
        _classCallCheck(this, TiBatteryHigh);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiBatteryHigh).apply(this, arguments));
    }

    _createClass(TiBatteryHigh, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm15 26.666666666666668c-0.9199999999999999 0-1.666666666666666-0.745000000000001-1.666666666666666-1.6666666666666679v-6.666666666666668c0-0.9216666666666669 0.7466666666666661-1.6666666666666679 1.666666666666666-1.6666666666666679s1.6666666666666679 0.745000000000001 1.6666666666666679 1.6666666666666679v6.666666666666668c0 0.9216666666666669-0.7466666666666661 1.6666666666666679-1.666666666666666 1.6666666666666679z m-5 0c-0.9199999999999999 0-1.666666666666666-0.745000000000001-1.666666666666666-1.6666666666666679v-6.666666666666668c0-0.9216666666666669 0.7466666666666661-1.6666666666666679 1.666666666666666-1.6666666666666679s1.666666666666666 0.745000000000001 1.666666666666666 1.6666666666666679v6.666666666666668c0 0.9216666666666669-0.7466666666666661 1.6666666666666679-1.666666666666666 1.6666666666666679z m10 0c-0.9200000000000017 0-1.6666666666666679-0.745000000000001-1.6666666666666679-1.6666666666666679v-6.666666666666668c0-0.9216666666666669 0.7466666666666661-1.6666666666666679 1.6666666666666679-1.6666666666666679s1.6666666666666679 0.745000000000001 1.6666666666666679 1.6666666666666679v6.666666666666668c0 0.9216666666666669-0.7466666666666661 1.6666666666666679-1.6666666666666679 1.6666666666666679z m11.666666666666668-10c0-2.7566666666666677-2.2433333333333323-5-5-5h-18.333333333333336c-2.756666666666665 0-4.999999999999998 2.243333333333334-4.999999999999998 5v10c0 2.7566666666666677 2.243333333333334 5 5 5h18.333333333333336c2.7566666666666677 0 5-2.2433333333333323 5-5 1.8400000000000034 0 3.3333333333333357-1.4933333333333323 3.3333333333333357-3.333333333333332v-3.333333333333332c0-1.8399999999999999-1.4933333333333323-3.333333333333332-3.333333333333332-3.333333333333332z m-3.333333333333332 10c0 0.9200000000000017-0.75 1.6666666666666679-1.6666666666666679 1.6666666666666679h-18.333333333333336c-0.9166666666666652 0-1.6666666666666652-0.7466666666666661-1.6666666666666652-1.6666666666666679v-10c0-0.9199999999999999 0.75-1.666666666666666 1.666666666666667-1.666666666666666h18.333333333333336c0.9166666666666679 0 1.6666666666666679 0.7466666666666661 1.6666666666666679 1.666666666666666v10z' })
                )
            );
        }
    }]);

    return TiBatteryHigh;
}(React.Component);

exports.default = TiBatteryHigh;
module.exports = exports['default'];