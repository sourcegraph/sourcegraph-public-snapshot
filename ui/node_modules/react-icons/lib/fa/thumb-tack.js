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

var FaThumbTack = function (_React$Component) {
    _inherits(FaThumbTack, _React$Component);

    function FaThumbTack() {
        _classCallCheck(this, FaThumbTack);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaThumbTack).apply(this, arguments));
    }

    _createClass(FaThumbTack, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm17.857142857142858 19.285714285714285v-10q0-0.31428571428571495-0.1999999999999993-0.5142857142857142t-0.514285714285716-0.1999999999999993-0.514285714285716 0.1999999999999993-0.1999999999999993 0.5142857142857142v10q0 0.31428571428571317 0.1999999999999993 0.5142857142857125t0.514285714285716 0.20000000000000284 0.514285714285716-0.1999999999999993 0.1999999999999993-0.514285714285716z m15.000000000000004 7.857142857142858q0 0.5800000000000018-0.42428571428571615 1.0042857142857144t-1.0042857142857144 0.42428571428571615h-9.575714285714284l-1.1385714285714315 10.781428571428574q-0.04285714285714448 0.2671428571428578-0.23428571428571487 0.4571428571428555t-0.4571428571428555 0.18999999999999773h-0.022857142857144908q-0.6028571428571432 0-0.7142857142857153-0.6028571428571396l-1.697142857142854-10.825714285714287h-9.01714285714286q-0.5800000000000001 0-1.0042857142857144-0.4242857142857126t-0.4242857142857135-1.004285714285718q0-2.7457142857142856 1.7528571428571427-4.942857142857143t3.9614285714285717-2.1999999999999993v-11.428571428571429q-1.1600000000000001 0-2.008571428571429-0.8485714285714288t-0.8485714285714288-2.008571428571428 0.8485714285714288-2.0085714285714285 2.008571428571429-0.8485714285714288h14.285714285714288q1.1600000000000001 0 2.008571428571429 0.8485714285714283t0.8485714285714252 2.008571428571429-0.8485714285714288 2.008571428571428-2.008571428571429 0.8485714285714288v11.428571428571429q2.210000000000001 0 3.9614285714285735 2.1999999999999993t1.7528571428571453 4.942857142857143z' })
                )
            );
        }
    }]);

    return FaThumbTack;
}(React.Component);

exports.default = FaThumbTack;
module.exports = exports['default'];