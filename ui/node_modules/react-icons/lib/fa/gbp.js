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

var FaGbp = function (_React$Component) {
    _inherits(FaGbp, _React$Component);

    function FaGbp() {
        _classCallCheck(this, FaGbp);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaGbp).apply(this, arguments));
    }

    _createClass(FaGbp, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm31.338571428571427 25.38v8.19142857142857q0 0.3142857142857167-0.1999999999999993 0.5142857142857125t-0.514285714285716 0.20000000000000284h-21.338571428571427q-0.31428571428571317 0-0.5142857142857125-0.20000000000000284t-0.20000000000000107-0.5142857142857125v-3.3485714285714288q0-0.28999999999999915 0.21000000000000085-0.5028571428571418t0.5028571428571436-0.21142857142856997h2.1642857142857146v-8.548571428571428h-2.120000000000001q-0.3114285714285714 0-0.5114285714285707-0.21142857142856997t-0.1999999999999993-0.5028571428571418v-2.9242857142857126q0-0.31428571428571317 0.1999999999999993-0.5142857142857125t0.5142857142857142-0.1999999999999993h2.1185714285714283v-4.978571428571435q0-3.814285714285714 2.757142857142858-6.292857142857144t7.021428571428569-2.478571428571429q4.12857142857143 0 7.475714285714286 2.7914285714285714 0.1999999999999993 0.17857142857142883 0.2228571428571442 0.4571428571428573t-0.15714285714285836 0.5028571428571427l-2.2985714285714245 2.8342857142857145q-0.1999999999999993 0.24571428571428555-0.48999999999999844 0.2671428571428578-0.28999999999999915 0.042857142857142705-0.514285714285716-0.15714285714285658-0.10999999999999943-0.10999999999999943-0.5785714285714292-0.4228571428571435t-1.5399999999999991-0.7142857142857135-2.0757142857142874-0.40000000000000036q-1.8957142857142841 0-3.057142857142857 1.048571428571428t-1.1600000000000001 2.7457142857142856v4.799999999999999h6.80857142857143q0.28999999999999915 0 0.5028571428571418 0.1999999999999993t0.21142857142856997 0.514285714285716v2.9228571428571435q0 0.28999999999999915-0.21142857142856997 0.5028571428571418t-0.5028571428571418 0.21142857142856997h-6.810000000000006v8.46h9.24285714285714v-4.039999999999999q0-0.28999999999999915 0.1999999999999993-0.5028571428571418t0.514285714285716-0.21142857142856997h3.614285714285714q0.31428571428571317 0 0.5142857142857125 0.21142857142856997t0.1999999999999993 0.5028571428571418z' })
                )
            );
        }
    }]);

    return FaGbp;
}(React.Component);

exports.default = FaGbp;
module.exports = exports['default'];