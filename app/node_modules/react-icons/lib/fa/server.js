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

var FaServer = function (_React$Component) {
    _inherits(FaServer, _React$Component);

    function FaServer() {
        _classCallCheck(this, FaServer);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaServer).apply(this, arguments));
    }

    _createClass(FaServer, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm2.857142857142857 31.42857142857143h22.857142857142858v-2.8571428571428577h-22.857142857142858v2.8571428571428577z m0-11.42857142857143h22.857142857142858v-2.8571428571428577h-22.857142857142858v2.8571428571428577z m35 10q0-0.8928571428571423-0.6257142857142881-1.5171428571428578t-1.5171428571428507-0.6257142857142846-1.5171428571428578 0.6257142857142846-0.6257142857142881 1.5171428571428578 0.6257142857142881 1.5171428571428578 1.5171428571428578 0.6257142857142881 1.5171428571428578-0.6257142857142846 0.6257142857142881-1.5171428571428613z m-35-21.42857142857143h22.85714285714286v-2.857142857142855h-22.857142857142858v2.8571428571428568z m35 10q0-0.8928571428571423-0.6257142857142881-1.5171428571428578t-1.5171428571428507-0.625714285714281-1.5171428571428578 0.6257142857142846-0.6257142857142881 1.5171428571428578 0.6257142857142881 1.5171428571428578 1.5171428571428578 0.6257142857142846 1.5171428571428578-0.6257142857142846 0.6257142857142881-1.5171428571428578z m0-11.428571428571429q0-0.8928571428571432-0.6257142857142881-1.5171428571428578t-1.5171428571428507-0.6257142857142828-1.5171428571428578 0.6257142857142854-0.6257142857142881 1.5171428571428578 0.6257142857142881 1.517142857142857 1.5171428571428578 0.6257142857142863 1.5171428571428578-0.6257142857142863 0.6257142857142881-1.517142857142857z m2.142857142857146 18.571428571428577v8.57142857142857h-40v-8.57142857142857h40z m0-11.428571428571429v8.571428571428571h-40v-8.571428571428571h40z m0-11.428571428571429v8.571428571428571h-40v-8.571428571428571h40z' })
                )
            );
        }
    }]);

    return FaServer;
}(React.Component);

exports.default = FaServer;
module.exports = exports['default'];