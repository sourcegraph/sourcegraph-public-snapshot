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

var FaDeviantart = function (_React$Component) {
    _inherits(FaDeviantart, _React$Component);

    function FaDeviantart() {
        _classCallCheck(this, FaDeviantart);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaDeviantart).apply(this, arguments));
    }

    _createClass(FaDeviantart, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm31.42857142857143 6.762857142857143l-6.762857142857143 12.991428571428571 0.5357142857142847 0.6914285714285704h6.227142857142859v9.262857142857143h-11.317142857142859l-0.9828571428571422 0.6714285714285708-3.171428571428571 6.092857142857145-0.668571428571429 0.6714285714285708h-6.717142857142859v-6.762857142857143l6.761428571428571-13.014285714285712-0.5357142857142865-0.668571428571429h-6.2285714285714295v-9.264285714285716h11.31857142857143l0.985714285714284-0.6714285714285717 3.1657142857142873-6.091428571428571 0.6714285714285708-0.6714285714285717h6.718571428571433v6.762857142857143z' })
                )
            );
        }
    }]);

    return FaDeviantart;
}(React.Component);

exports.default = FaDeviantart;
module.exports = exports['default'];