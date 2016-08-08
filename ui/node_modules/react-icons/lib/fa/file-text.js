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

var FaFileText = function (_React$Component) {
    _inherits(FaFileText, _React$Component);

    function FaFileText() {
        _classCallCheck(this, FaFileText);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaFileText).apply(this, arguments));
    }

    _createClass(FaFileText, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm35.62428571428572 10.625714285714286q0.3142857142857167 0.31428571428571495 0.6257142857142881 0.8028571428571425h-10.535714285714292v-10.535714285714286q0.4914285714285711 0.31428571428571495 0.8028571428571425 0.6257142857142863z m-10.624285714285719 3.66h12.142857142857146v23.571428571428577q0 0.8928571428571459-0.6257142857142881 1.5171428571428578t-1.5171428571428578 0.625714285714281h-30q-0.8928571428571432 0-1.5171428571428573-0.6257142857142881t-0.6257142857142854-1.5171428571428507v-35.714285714285715q0-0.8928571428571459 0.6257142857142859-1.51714285714286t1.517142857142857-0.6257142857142859h17.857142857142858v12.142857142857142q0 0.8928571428571423 0.6257142857142846 1.5171428571428578t1.5171428571428578 0.6257142857142863z m3.571428571428573 16.42857142857143v-1.428571428571427q0-0.31428571428571317-0.1999999999999993-0.514285714285716t-0.514285714285716-0.1999999999999993h-15.714285714285715q-0.31428571428571495 0-0.5142857142857142 0.1999999999999993t-0.1999999999999993 0.5142857142857125v1.428571428571427q0 0.31428571428571317 0.1999999999999993 0.514285714285716t0.5142857142857142 0.1999999999999993h15.714285714285715q0.31428571428571317 0 0.514285714285716-0.1999999999999993t0.1999999999999993-0.5142857142857125z m0-5.714285714285715v-1.428571428571427q0-0.31428571428571317-0.1999999999999993-0.514285714285716t-0.514285714285716-0.1999999999999993h-15.714285714285715q-0.31428571428571495 0-0.5142857142857142 0.1999999999999993t-0.1999999999999993 0.514285714285716v1.428571428571427q0 0.31428571428571317 0.1999999999999993 0.5142857142857125t0.5142857142857142 0.1999999999999993h15.714285714285715q0.31428571428571317 0 0.514285714285716-0.1999999999999993t0.1999999999999993-0.5142857142857125z m0-5.714285714285715v-1.428571428571427q0-0.31428571428571317-0.1999999999999993-0.514285714285716t-0.514285714285716-0.1999999999999993h-15.714285714285715q-0.31428571428571495 0-0.5142857142857142 0.1999999999999993t-0.1999999999999993 0.514285714285716v1.428571428571427q0 0.31428571428571317 0.1999999999999993 0.5142857142857125t0.5142857142857142 0.1999999999999993h15.714285714285715q0.31428571428571317 0 0.514285714285716-0.1999999999999993t0.1999999999999993-0.5142857142857125z' })
                )
            );
        }
    }]);

    return FaFileText;
}(React.Component);

exports.default = FaFileText;
module.exports = exports['default'];