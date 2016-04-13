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

var MdViewWeek = function (_React$Component) {
    _inherits(MdViewWeek, _React$Component);

    function MdViewWeek() {
        _classCallCheck(this, MdViewWeek);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdViewWeek).apply(this, arguments));
    }

    _createClass(MdViewWeek, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm21.64 8.360000000000001q0.7033333333333331 0 1.211666666666666 0.4666666666666668t0.5100000000000016 1.1733333333333338v20q0 0.7033333333333331-0.5083333333333329 1.1716666666666669t-1.2100000000000009 0.466666666666665h-5q-0.7033333333333331 0-1.1716666666666669-0.466666666666665t-0.47166666666666757-1.1716666666666669v-20q0-0.7033333333333331 0.4666666666666668-1.1716666666666669t1.1733333333333338-0.4666666666666668h5z m11.719999999999999 0q0.7033333333333331 0 1.1716666666666669 0.4666666666666668t0.4683333333333337 1.173333333333332v20q0 0.7033333333333331-0.46666666666666856 1.1716666666666669t-1.173333333333332 0.466666666666665h-5q-0.7033333333333331 0-1.211666666666666-0.466666666666665t-0.509999999999998-1.1716666666666669v-20q0-0.7033333333333331 0.5083333333333329-1.1716666666666669t1.2100000000000009-0.4666666666666668h5z m-23.36 0q0.7033333333333331 0 1.1716666666666669 0.4666666666666668t0.4666666666666668 1.1733333333333338v20q0 0.7033333333333331-0.4666666666666668 1.1716666666666669t-1.1716666666666669 0.466666666666665h-5q-0.7033333333333331 0-1.1716666666666669-0.466666666666665t-0.46666666666666634-1.1716666666666669v-20q0-0.7033333333333331 0.4666666666666668-1.1716666666666669t1.1716666666666664-0.466666666666665h5z' })
                )
            );
        }
    }]);

    return MdViewWeek;
}(React.Component);

exports.default = MdViewWeek;
module.exports = exports['default'];