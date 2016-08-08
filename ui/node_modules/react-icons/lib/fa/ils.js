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

var FaIls = function (_React$Component) {
    _inherits(FaIls, _React$Component);

    function FaIls() {
        _classCallCheck(this, FaIls);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaIls).apply(this, arguments));
    }

    _createClass(FaIls, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm27.142857142857142 13.928571428571429v11.071428571428571q0 0.31428571428571317-0.1999999999999993 0.5142857142857125t-0.5142857142857125 0.20000000000000284h-3.571428571428573q-0.31428571428571317 0-0.514285714285716-0.1999999999999993t-0.1999999999999993-0.514285714285716v-11.071428571428571q0-2.5-1.7857142857142847-4.2857142857142865t-4.285714285714285-1.7857142857142856h-6.071428571428573v25.71428571428571q0 0.3142857142857167-0.1999999999999993 0.5142857142857125t-0.5142857142857142 0.20000000000000284h-3.571428571428572q-0.31428571428571406 0-0.5142857142857142-0.20000000000000284t-0.20000000000000018-0.5142857142857125v-30q0-0.31428571428571406 0.20000000000000018-0.5142857142857142t0.5142857142857142-0.19999999999999796h10.357142857142858q3.014285714285716 0 5.557142857142857 1.4857142857142862t4.0285714285714285 4.0285714285714285 1.4857142857142875 5.557142857142859z m8.571428571428573-10.357142857142858v19.642857142857146q0 3.0142857142857125-1.4857142857142875 5.557142857142857t-4.0285714285714285 4.028571428571425-5.557142857142857 1.4857142857142875h-10.357142857142856q-0.31428571428571495 0-0.5142857142857142-0.20000000000000284t-0.20000000000000107-0.5142857142857125v-21.42857142857143q0-0.31428571428571495 0.1999999999999993-0.5142857142857142t0.514285714285716-0.19999999999999574h3.571428571428571q0.31428571428571317 0 0.514285714285716 0.1999999999999993t0.1999999999999993 0.5142857142857142v17.142857142857142h6.071428571428573q2.5 0 4.285714285714285-1.7857142857142847t1.7857142857142847-4.285714285714285v-19.642857142857142q0-0.31428571428571583 0.1999999999999993-0.5142857142857156t0.514285714285716-0.20000000000000018h3.5714285714285694q0.3142857142857167 0 0.5142857142857125 0.20000000000000018t0.20000000000000284 0.5142857142857142z' })
                )
            );
        }
    }]);

    return FaIls;
}(React.Component);

exports.default = FaIls;
module.exports = exports['default'];