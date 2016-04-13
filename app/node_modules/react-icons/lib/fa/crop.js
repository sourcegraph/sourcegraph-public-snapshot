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

var FaCrop = function (_React$Component) {
    _inherits(FaCrop, _React$Component);

    function FaCrop() {
        _classCallCheck(this, FaCrop);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaCrop).apply(this, arguments));
    }

    _createClass(FaCrop, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm13.86142857142857 28.571428571428573h13.281428571428572v-13.281428571428572z m-1.0042857142857127-1.0042857142857144l13.281428571428574-13.281428571428572h-13.281428571428574v13.281428571428572z m25.71428571428571 1.7185714285714262v4.285714285714285q0 0.3142857142857167-0.20000000000000284 0.5142857142857125t-0.5142857142857125 0.20000000000000284h-4.999999999999993v5q0 0.3142857142857167-0.20000000000000284 0.5142857142857125t-0.5142857142857125 0.20000000000000284h-4.285714285714285q-0.31428571428571317 0-0.5142857142857125-0.20000000000000284t-0.2000000000000064-0.5142857142857125v-5h-19.285714285714285q-0.31428571428571406 0-0.5142857142857142-0.20000000000000284t-0.20000000000000018-0.5142857142857125v-19.285714285714285h-5q-0.3142857142857147 1.7763568394002505e-15-0.5142857142857147-0.1999999999999975t-0.19999999999999996-0.514285714285716v-4.2857142857142865q0-0.31428571428571495 0.19999999999999996-0.5142857142857142t0.5142857142857142-0.1999999999999993h5v-5q8.881784197001252e-16-0.31428571428571406 0.20000000000000107-0.5142857142857138t0.5142857142857142-0.20000000000000018h4.285714285714285q0.31428571428571495 0 0.5142857142857142 0.20000000000000018t0.20000000000000107 0.5142857142857142v5h18.995714285714286l5.491428571428575-5.514285714285714q0.2242857142857133-0.20000000000000018 0.5142857142857125-0.20000000000000018t0.5142857142857125 0.20000000000000018q0.20000000000000284 0.2242857142857142 0.20000000000000284 0.5142857142857142t-0.20000000000000284 0.5142857142857142l-5.515714285714282 5.488571428571429v18.997142857142858h5q0.3142857142857167 0 0.5142857142857125 0.1999999999999993t0.20000000000000284 0.5142857142857125z' })
                )
            );
        }
    }]);

    return FaCrop;
}(React.Component);

exports.default = FaCrop;
module.exports = exports['default'];