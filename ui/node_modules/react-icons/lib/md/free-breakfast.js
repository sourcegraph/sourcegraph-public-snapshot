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

var MdFreeBreakfast = function (_React$Component) {
    _inherits(MdFreeBreakfast, _React$Component);

    function MdFreeBreakfast() {
        _classCallCheck(this, MdFreeBreakfast);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdFreeBreakfast).apply(this, arguments));
    }

    _createClass(MdFreeBreakfast, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm6.640000000000001 31.640000000000004h26.71666666666667v3.359999999999996h-26.715000000000003v-3.3599999999999994z m26.72-18.28v-5.000000000000002h-3.3599999999999994v5h3.3599999999999994z m0-8.360000000000003q1.4066666666666663 0 2.3433333333333337 0.9766666666666666t0.9383333333333326 2.383333333333333v5q0 1.3283333333333331-0.9383333333333326 2.3049999999999997t-2.3433333333333337 0.9750000000000014h-3.3599999999999994v5q0 2.7333333333333343-1.9533333333333331 4.726666666666667t-4.688333333333333 1.9933333333333323h-10q-2.7333333333333343 0-4.726666666666667-1.9916666666666671t-1.9933333333333332-4.725000000000001v-16.64333333333333h26.71666666666667z' })
                )
            );
        }
    }]);

    return MdFreeBreakfast;
}(React.Component);

exports.default = MdFreeBreakfast;
module.exports = exports['default'];