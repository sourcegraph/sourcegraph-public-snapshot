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

var FaSignal = function (_React$Component) {
    _inherits(FaSignal, _React$Component);

    function FaSignal() {
        _classCallCheck(this, FaSignal);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaSignal).apply(this, arguments));
    }

    _createClass(FaSignal, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm5.714285714285714 32.142857142857146v4.285714285714285q0 0.3142857142857167-0.20000000000000018 0.5142857142857125t-0.5142857142857142 0.20000000000000284h-4.285714285714286q-0.3142857142857143 0-0.5142857142857142-0.20000000000000284t-0.20000000000000018-0.5142857142857125v-4.285714285714285q0-0.31428571428571317 0.2-0.5142857142857125t0.5142857142857142-0.20000000000000284h4.285714285714286q0.31428571428571406 0 0.5142857142857142 0.1999999999999993t0.20000000000000018 0.514285714285716z m8.57142857142857-2.8571428571428577v7.142857142857142q0 0.3142857142857167-0.1999999999999993 0.5142857142857125t-0.5142857142857142 0.20000000000000284h-4.2857142857142865q-0.31428571428571495 0-0.5142857142857142-0.20000000000000284t-0.1999999999999993-0.5142857142857125v-7.142857142857142q0-0.31428571428571317 0.1999999999999993-0.514285714285716t0.514285714285716-0.1999999999999993h4.2857142857142865q0.31428571428571495 0 0.5142857142857142 0.1999999999999993t0.1999999999999993 0.5142857142857125z m8.57142857142857-5.714285714285715v12.857142857142858q0 0.3142857142857167-0.1999999999999993 0.5142857142857125t-0.514285714285716 0.20000000000000284h-4.285714285714285q-0.31428571428571317 0-0.5142857142857125-0.20000000000000284t-0.1999999999999993-0.5142857142857125v-12.857142857142858q0-0.31428571428571317 0.1999999999999993-0.514285714285716t0.5142857142857125-0.1999999999999993h4.285714285714285q0.31428571428571317 0 0.5142857142857125 0.1999999999999993t0.1999999999999993 0.5142857142857125z m8.57142857142857-8.571428571428571v21.42857142857143q0 0.3142857142857167-0.1999999999999993 0.5142857142857125t-0.514285714285716 0.20000000000000284h-4.285714285714285q-0.31428571428571317 0-0.5142857142857125-0.20000000000000284t-0.19999999999999574-0.5142857142857125v-21.42857142857143q0-0.31428571428571495 0.1999999999999993-0.5142857142857142t0.514285714285716-0.1999999999999993h4.285714285714285q0.31428571428571317 0 0.5142857142857125 0.1999999999999993t0.1999999999999993 0.5142857142857142z m8.57142857142857-11.428571428571429v32.85714285714286q0 0.3142857142857167-0.20000000000000284 0.5142857142857125t-0.5142857142857125 0.20000000000000284h-4.285714285714285q-0.3142857142857167 0-0.5142857142857125-0.20000000000000284t-0.19999999999999574-0.5142857142857125v-32.85714285714286q0-0.31428571428571406 0.20000000000000284-0.5142857142857142t0.5142857142857125-0.19999999999999796h4.285714285714285q0.3142857142857167 0 0.5142857142857125 0.20000000000000018t0.20000000000000284 0.5142857142857142z' })
                )
            );
        }
    }]);

    return FaSignal;
}(React.Component);

exports.default = FaSignal;
module.exports = exports['default'];