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

var TiWeatherWindy = function (_React$Component) {
    _inherits(TiWeatherWindy, _React$Component);

    function TiWeatherWindy() {
        _classCallCheck(this, TiWeatherWindy);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiWeatherWindy).apply(this, arguments));
    }

    _createClass(TiWeatherWindy, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm31.666666666666668 8.333333333333334c-0.9216666666666669 0-1.6666666666666679 0.7449999999999992-1.6666666666666679 1.666666666666666s0.745000000000001 1.666666666666666 1.6666666666666679 1.666666666666666c0.9199999999999982 0 1.6666666666666679 0.7466666666666661 1.6666666666666679 1.666666666666666s-0.7466666666666697 1.666666666666666-1.6666666666666679 1.666666666666666h-18.333333333333336c-0.9216666666666651 0-1.6666666666666643 0.7449999999999992-1.6666666666666643 1.666666666666666s0.7449999999999992 1.6666666666666679 1.666666666666666 1.6666666666666679h10.000000000000002c0.9200000000000017 0 1.6666666666666679 0.7466666666666661 1.6666666666666679 1.6666666666666679s-0.7466666666666661 1.6666666666666679-1.6666666666666679 1.6666666666666679h-10.66666666666667c-2.756666666666666 0-5 2.2433333333333323-5 5s2.243333333333334 5 5 5c0.9216666666666669 0 1.666666666666666-0.745000000000001 1.666666666666666-1.6666666666666679s-0.7449999999999992-1.6666666666666679-1.666666666666666-1.6666666666666679c-0.9199999999999999 0-1.666666666666666-0.7466666666666661-1.666666666666666-1.6666666666666679s0.7466666666666661-1.6666666666666679 1.666666666666666-1.6666666666666679h10.66666666666667c2.7566666666666677 0 5-2.2433333333333323 5-5 0-0.5883333333333347-0.120000000000001-1.1433333333333344-0.30833333333333357-1.6666666666666679h3.6416666666666657c2.756666666666664 0 5.0000000000000036-2.2433333333333323 5.0000000000000036-5s-2.2433333333333323-5-5-5z' })
                )
            );
        }
    }]);

    return TiWeatherWindy;
}(React.Component);

exports.default = TiWeatherWindy;
module.exports = exports['default'];