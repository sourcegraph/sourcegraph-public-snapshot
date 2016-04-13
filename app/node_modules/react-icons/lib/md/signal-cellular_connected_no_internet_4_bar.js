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

var MdSignalCellularConnectedNoInternet4Bar = function (_React$Component) {
    _inherits(MdSignalCellularConnectedNoInternet4Bar, _React$Component);

    function MdSignalCellularConnectedNoInternet4Bar() {
        _classCallCheck(this, MdSignalCellularConnectedNoInternet4Bar);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdSignalCellularConnectedNoInternet4Bar).apply(this, arguments));
    }

    _createClass(MdSignalCellularConnectedNoInternet4Bar, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm3.3600000000000003 36.64000000000001l33.28333333333333-33.28333333333333v10h-6.643333333333331v23.28333333333334h-26.636666666666667z m30 0v-3.2833333333333314h3.2833333333333314v3.2833333333333314h-3.2833333333333314z m0-6.640000000000001v-13.360000000000007h3.2833333333333314v13.36h-3.2833333333333314z' })
                )
            );
        }
    }]);

    return MdSignalCellularConnectedNoInternet4Bar;
}(React.Component);

exports.default = MdSignalCellularConnectedNoInternet4Bar;
module.exports = exports['default'];