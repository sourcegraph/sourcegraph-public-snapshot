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

var MdDiscFull = function (_React$Component) {
    _inherits(MdDiscFull, _React$Component);

    function MdDiscFull() {
        _classCallCheck(this, MdDiscFull);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdDiscFull).apply(this, arguments));
    }

    _createClass(MdDiscFull, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm16.64 23.36q1.3283333333333331 0 2.3433333333333337-1.0166666666666657t1.0166666666666657-2.3416666666666686-1.0166666666666657-2.3416666666666686-2.3433333333333337-1.0166666666666657-2.3049999999999997 1.0166666666666657-0.9749999999999996 2.3400000000000034 0.9766666666666666 2.344999999999999 2.3066666666666666 1.0166666666666657z m0-16.720000000000002q5.546666666666667 0 9.453333333333333 3.9450000000000003t3.9066666666666663 9.415000000000003-3.9066666666666663 9.411666666666669-9.453333333333333 3.9450000000000003q-5.466666666666667 0-9.375-3.9066666666666663t-3.9083333333333337-9.453333333333333 3.9100000000000006-9.453333333333337 9.373333333333333-3.9066666666666663z m16.72 5h3.2833333333333314v8.360000000000003h-3.2833333333333314v-8.36z m0 15v-3.283333333333335h3.2833333333333314v3.2833333333333314h-3.2833333333333314z' })
                )
            );
        }
    }]);

    return MdDiscFull;
}(React.Component);

exports.default = MdDiscFull;
module.exports = exports['default'];