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

var FaRouble = function (_React$Component) {
    _inherits(FaRouble, _React$Component);

    function FaRouble() {
        _classCallCheck(this, FaRouble);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaRouble).apply(this, arguments));
    }

    _createClass(FaRouble, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm28.995714285714286 12.61142857142857q0-2.232857142857142-1.451428571428572-3.614285714285714t-3.814285714285713-1.3857142857142861h-7.142857142857142v9.999999999999998h7.142857142857142q2.364285714285714 0 3.814285714285713-1.3857142857142861t1.452857142857141-3.614285714285714z m5.289999999999999 0q0 4.308571428571428-2.8242857142857147 7.031428571428572t-7.28857142857143 2.722857142857144h-7.5871428571428545v2.6342857142857135h11.271428571428569q0.31428571428571317 0 0.5142857142857125 0.1999999999999993t0.1999999999999993 0.514285714285716v2.8571428571428577q0 0.31428571428571317-0.1999999999999993 0.514285714285716t-0.5142857142857125 0.1999999999999993h-11.271428571428569v4.285714285714288q0 0.3142857142857167-0.21142857142856997 0.5142857142857125t-0.5028571428571436 0.20000000000000284h-3.7285714285714295q-0.3114285714285714 0-0.5114285714285707-0.20000000000000284t-0.20285714285714285-0.5142857142857196v-4.285714285714285h-5q-0.31428571428571406 0-0.5142857142857142-0.1999999999999993t-0.20000000000000018-0.5142857142857125v-2.8571428571428577q0-0.31428571428571317 0.20000000000000018-0.5142857142857125t0.5142857142857142-0.1999999999999993h5v-2.6342857142857135h-5q-0.31428571428571406 0-0.5142857142857142-0.1999999999999993t-0.20000000000000018-0.514285714285716v-3.325714285714284q0-0.28999999999999915 0.20000000000000018-0.5028571428571418t0.5142857142857142-0.21142857142856997h5v-14.040000000000008q0-0.31428571428571406 0.1999999999999993-0.5142857142857138t0.5142857142857142-0.20000000000000018h12.028571428571428q4.465714285714284 0 7.289999999999999 2.722857142857143t2.8242857142857147 7.03142857142857z' })
                )
            );
        }
    }]);

    return FaRouble;
}(React.Component);

exports.default = FaRouble;
module.exports = exports['default'];