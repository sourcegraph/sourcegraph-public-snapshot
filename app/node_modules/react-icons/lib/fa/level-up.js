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

var FaLevelUp = function (_React$Component) {
    _inherits(FaLevelUp, _React$Component);

    function FaLevelUp() {
        _classCallCheck(this, FaLevelUp);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaLevelUp).apply(this, arguments));
    }

    _createClass(FaLevelUp, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm31.294285714285714 13.46q-0.40000000000000213 0.8257142857142856-1.2942857142857136 0.8257142857142856h-4.285714285714285v19.285714285714285q0 0.3142857142857167-0.1999999999999993 0.5142857142857125t-0.514285714285716 0.20000000000000284h-15.714285714285714q-0.468571428571428 0-0.6471428571428568-0.3999999999999986-0.17857142857142883-0.4485714285714266 0.08999999999999986-0.7828571428571394l3.571428571428571-4.285714285714285q0.1999999999999993-0.2457142857142891 0.5571428571428569-0.2457142857142891h7.142857142857142v-14.285714285714286h-4.285714285714285q-0.8914285714285715 0-1.2928571428571427-0.8257142857142856-0.3800000000000008-0.8257142857142856 0.1999999999999993-1.5171428571428578l7.142857142857144-8.571428571428571q0.4028571428571439-0.49142857142857155 1.095714285714287-0.49142857142857155t1.0942857142857143 0.49142857142857155l7.142857142857142 8.571428571428571q0.6028571428571432 0.7142857142857135 0.1999999999999993 1.5171428571428578z' })
                )
            );
        }
    }]);

    return FaLevelUp;
}(React.Component);

exports.default = FaLevelUp;
module.exports = exports['default'];