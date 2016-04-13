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

var TiRadar = function (_React$Component) {
    _inherits(TiRadar, _React$Component);

    function TiRadar() {
        _classCallCheck(this, TiRadar);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiRadar).apply(this, arguments));
    }

    _createClass(TiRadar, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm20 33.333333333333336c6.433333333333334 0 11.666666666666668-5.233333333333334 11.666666666666668-11.666666666666668s-5.233333333333334-11.666666666666668-11.671666666666667-11.666666666666668c-6.428333333333333 0-11.661666666666667 5.233333333333334-11.661666666666667 11.666666666666668s5.2333333333333325 11.666666666666668 11.666666666666666 11.666666666666668z m-1.6666666666666643-19.830000000000002v3.163333333333334c0 0.9216666666666669 0.7466666666666661 1.6666666666666679 1.6666666666666679 1.6666666666666679s1.6666666666666679-0.745000000000001 1.6666666666666679-1.6666666666666679v-3.163333333333334c3.2600000000000016 0.663333333333334 5.833333333333336 3.236666666666668 6.5 6.496666666666666h-3.1666666666666714c-0.9200000000000017 0-1.6666666666666679 0.745000000000001-1.6666666666666679 1.6666666666666679s0.7466666666666661 1.6666666666666679 1.6666666666666679 1.6666666666666679h3.166666666666668c-0.6666666666666679 3.2600000000000016-3.2399999999999984 5.833333333333336-6.5 6.496666666666666v-3.163333333333334c0-0.9216666666666669-0.7466666666666661-1.6666666666666679-1.6666666666666679-1.6666666666666679s-1.6666666666666679 0.745000000000001-1.6666666666666679 1.6666666666666679v3.163333333333334c-3.259999999999998-0.663333333333334-5.833333333333332-3.236666666666668-6.499999999999998-6.496666666666666h3.166666666666666c0.9199999999999999 0 1.6666666666666679-0.745000000000001 1.6666666666666679-1.6666666666666679s-0.7466666666666661-1.6666666666666679-1.666666666666666-1.6666666666666679h-3.166666666666668c0.6666666666666661-3.2600000000000016 3.2383333333333333-5.833333333333334 6.500000000000002-6.496666666666666z' })
                )
            );
        }
    }]);

    return TiRadar;
}(React.Component);

exports.default = TiRadar;
module.exports = exports['default'];