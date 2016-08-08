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

var FaMobile = function (_React$Component) {
    _inherits(FaMobile, _React$Component);

    function FaMobile() {
        _classCallCheck(this, FaMobile);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaMobile).apply(this, arguments));
    }

    _createClass(FaMobile, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm21.785714285714285 31.42857142857143q0-0.7371428571428567-0.524285714285714-1.2614285714285707t-1.2614285714285707-0.5242857142857176-1.2614285714285707 0.524285714285714-0.524285714285714 1.2614285714285742 0.524285714285714 1.2614285714285742 1.2614285714285707 0.5242857142857176 1.2614285714285707-0.5242857142857176 0.524285714285714-1.2614285714285742z m4.642857142857142-3.571428571428573v-15.714285714285715q0-0.28999999999999915-0.21142857142856997-0.5028571428571436t-0.5028571428571418-0.21142857142856997h-11.428571428571429q-0.2900000000000009 0-0.5028571428571436 0.21142857142857174t-0.21142857142857174 0.5028571428571418v15.714285714285715q0 0.28999999999999915 0.21142857142857174 0.5028571428571418t0.5028571428571436 0.21142857142857352h11.428571428571429q0.28999999999999915 0 0.5028571428571418-0.21142857142856997t0.21142857142856997-0.5028571428571418z m-4.285714285714285-18.92857142857143q0-0.3571428571428559-0.35714285714285765-0.3571428571428559h-3.571428571428573q-0.35714285714285765 0-0.35714285714285765 0.35714285714285765t0.35714285714285765 0.35714285714285765h3.571428571428573q0.35714285714285765 0 0.35714285714285765-0.35714285714285765z m6.428571428571431-0.3571428571428559v22.85714285714286q0 1.1599999999999966-0.8485714285714288 2.008571428571429t-2.008571428571429 0.8485714285714252h-11.428571428571429q-1.1600000000000001 0-2.008571428571429-0.8485714285714252t-0.8485714285714288-2.008571428571429v-22.85714285714286q0-1.1599999999999984 0.8485714285714288-2.008571428571427t2.008571428571429-0.8485714285714279h11.428571428571429q1.1600000000000001 0 2.008571428571429 0.8485714285714288t0.8485714285714288 2.008571428571428z' })
                )
            );
        }
    }]);

    return FaMobile;
}(React.Component);

exports.default = FaMobile;
module.exports = exports['default'];