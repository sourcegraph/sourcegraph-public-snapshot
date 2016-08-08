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

var MdWatch = function (_React$Component) {
    _inherits(MdWatch, _React$Component);

    function MdWatch() {
        _classCallCheck(this, MdWatch);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdWatch).apply(this, arguments));
    }

    _createClass(MdWatch, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm10 20q0 4.140000000000001 2.9299999999999997 7.07t7.07 2.9299999999999997 7.07-2.9299999999999997 2.9299999999999997-7.07-2.9299999999999997-7.07-7.07-2.9299999999999997-7.07 2.9299999999999997-2.9299999999999997 7.07z m23.36 0q0 6.406666666666666-5.078333333333333 10.466666666666669l-1.6416666666666657 9.533333333333331h-13.283333333333333l-1.6383333333333336-9.533333333333331q-5.076666666666667-3.9033333333333324-5.076666666666667-10.466666666666669t5.078333333333334-10.466666666666667l1.6400000000000006-9.533333333333333h13.283333333333333l1.6383333333333319 9.533333333333333q5.076666666666668 4.061666666666666 5.076666666666668 10.466666666666667z' })
                )
            );
        }
    }]);

    return MdWatch;
}(React.Component);

exports.default = MdWatch;
module.exports = exports['default'];