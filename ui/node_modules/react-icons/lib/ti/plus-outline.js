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

var TiPlusOutline = function (_React$Component) {
    _inherits(TiPlusOutline, _React$Component);

    function TiPlusOutline() {
        _classCallCheck(this, TiPlusOutline);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiPlusOutline).apply(this, arguments));
    }

    _createClass(TiPlusOutline, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm20 35c-2.7566666666666677 0-5-2.2433333333333323-5-5l0.08833333333333293-5.088333333333335-5.058333333333334 0.08833333333333471c-2.7866666666666653 0-5.029999999999999-2.2433333333333323-5.029999999999999-5s2.243333333333334-5 5-5l5.088333333333335-0.08999999999999986-0.08833333333333471-4.880000000000001c0-2.7866666666666653 2.243333333333336-5.029999999999999 5-5.029999999999999s5 2.243333333333334 5 5l0.09166666666666501 4.91 4.938333333333336 0.08999999999999986c2.7266666666666666 0 4.969999999999999 2.243333333333336 4.969999999999999 5s-2.2433333333333323 5-5 5l-4.908333333333331-0.08833333333333471-0.09166666666666501 5.116666666666667c-3.552713678800501e-15 2.7300000000000075-2.243333333333336 4.971666666666668-5.0000000000000036 4.971666666666668z m-1.6666666666666679-13.333333333333336v8.363333333333333c0 0.8883333333333319 0.75 1.6366666666666667 1.6666666666666679 1.6366666666666667s1.6666666666666679-0.75 1.6666666666666679-1.6666666666666679v-8.333333333333332h8.363333333333333c0.8883333333333319 0 1.6366666666666667-0.75 1.6366666666666667-1.6666666666666679s-0.75-1.6666666666666679-1.6666666666666679-1.6666666666666679h-8.333333333333336v-8.333333333333329c0-0.9499999999999993-0.75-1.666666666666666-1.6666666666666679-1.666666666666666s-1.6666666666666679 0.75-1.6666666666666679 1.666666666666666v8.333333333333336h-8.333333333333329c-0.9499999999999993 0-1.666666666666666 0.75-1.666666666666666 1.6666666666666679s0.75 1.6666666666666679 1.666666666666666 1.6666666666666679h8.333333333333336z' })
                )
            );
        }
    }]);

    return TiPlusOutline;
}(React.Component);

exports.default = TiPlusOutline;
module.exports = exports['default'];