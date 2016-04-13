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

var MdFormatBold = function (_React$Component) {
    _inherits(MdFormatBold, _React$Component);

    function MdFormatBold() {
        _classCallCheck(this, MdFormatBold);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdFormatBold).apply(this, arguments));
    }

    _createClass(MdFormatBold, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm22.5 25.86q1.0933333333333337 0 1.7966666666666669-0.7416666666666671t0.7033333333333331-1.7566666666666642-0.7033333333333331-1.7583333333333329-1.7966666666666669-0.7399999999999984h-5.859999999999999v5h5.859999999999999z m-5.859999999999999-15v5h5q1.0166666666666657 0 1.7583333333333329-0.7416666666666671t0.7399999999999984-1.7566666666666677-0.7416666666666671-1.7583333333333329-1.7600000000000016-0.7400000000000002h-5z m9.375 7.108333333333334q3.591666666666665 1.6416666666666657 3.591666666666665 5.704999999999998 0 2.6566666666666663-1.7583333333333329 4.491666666666667t-4.411666666666665 1.8350000000000009h-11.796666666666667v-23.36h10.466666666666669q2.8166666666666664 0 4.728333333333332 1.9533333333333331t1.9166666666666679 4.766666666666666-2.736666666666668 4.608333333333334z' })
                )
            );
        }
    }]);

    return MdFormatBold;
}(React.Component);

exports.default = MdFormatBold;
module.exports = exports['default'];