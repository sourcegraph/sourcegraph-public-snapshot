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

var MdLocalCafe = function (_React$Component) {
    _inherits(MdLocalCafe, _React$Component);

    function MdLocalCafe() {
        _classCallCheck(this, MdLocalCafe);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdLocalCafe).apply(this, arguments));
    }

    _createClass(MdLocalCafe, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm3.3600000000000003 35v-3.3599999999999994h30v3.3599999999999994h-30z m30-21.64v-5h-3.3599999999999994v5h3.3599999999999994z m0-8.360000000000001q1.4066666666666663 0 2.3433333333333337 0.9766666666666666t0.9383333333333326 2.383333333333333v5q0 1.4066666666666663-0.9383333333333326 2.3433333333333337t-2.3433333333333337 0.9383333333333344h-3.3599999999999994v5q0 2.7333333333333343-1.9533333333333331 4.726666666666667t-4.688333333333333 1.9933333333333323h-10q-2.7333333333333343 0-4.726666666666667-1.9916666666666671t-1.9933333333333332-4.725000000000001v-16.644999999999996h26.71666666666667z' })
                )
            );
        }
    }]);

    return MdLocalCafe;
}(React.Component);

exports.default = MdLocalCafe;
module.exports = exports['default'];