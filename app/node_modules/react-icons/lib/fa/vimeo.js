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

var FaVimeo = function (_React$Component) {
    _inherits(FaVimeo, _React$Component);

    function FaVimeo() {
        _classCallCheck(this, FaVimeo);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaVimeo).apply(this, arguments));
    }

    _createClass(FaVimeo, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm38.14714285714286 11.562857142857142q-0.2228571428571442 5.267142857142856-7.41 14.531428571428572-7.432857142857149 9.620000000000001-12.542857142857144 9.620000000000001-3.171428571428571 0-5.357142857142858-5.871428571428574-0.9857142857142858-3.571428571428573-2.9485714285714284-10.75714285714286-1.6071428571428577-5.848571428571425-3.5028571428571444-5.848571428571425-0.402857142857143 0-2.835714285714286 1.6971428571428575l-1.7185714285714286-2.185714285714287q0.5357142857142858-0.47142857142857153 2.41-2.1571428571428566t2.901428571428572-2.5771428571428565q3.482857142857143-3.08 5.38-3.257142857142857 2.119999999999999-0.20285714285714285 3.4142857142857146 1.2371428571428575t1.8100000000000005 4.542857142857143q0.9828571428571422 6.405714285714286 1.4714285714285715 8.325714285714287 1.2285714285714278 5.557142857142857 2.6799999999999997 5.557142857142857 1.1371428571428588 0 3.435714285714287-3.5928571428571416 2.2542857142857144-3.595714285714287 2.4328571428571415-5.492857142857144 0.28999999999999915-3.102857142857143-2.4328571428571415-3.102857142857143-1.2714285714285722 0-2.6999999999999993 0.5800000000000001 2.677142857142858-8.774285714285716 10.24285714285714-8.528571428571428 5.604285714285716 0.17857142857142883 5.268571428571427 7.277142857142858z' })
                )
            );
        }
    }]);

    return FaVimeo;
}(React.Component);

exports.default = FaVimeo;
module.exports = exports['default'];