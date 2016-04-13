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

var FaMars = function (_React$Component) {
    _inherits(FaMars, _React$Component);

    function FaMars() {
        _classCallCheck(this, FaMars);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaMars).apply(this, arguments));
    }

    _createClass(FaMars, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm35.714285714285715 2.857142857142857q0.5799999999999983 0 1.0042857142857144 0.4242857142857144t0.42428571428571615 1.004285714285714v9.285714285714288q0 0.31428571428571495-0.20000000000000284 0.5142857142857142t-0.5142857142857125 0.1999999999999993h-1.4285714285714306q-0.3142857142857167 0-0.5142857142857125-0.1999999999999993t-0.20000000000000284-0.514285714285716v-5.848571428571429l-8.528571428571428 8.54857142857143q2.8142857142857167 3.482857142857142 2.8142857142857167 8.014285714285712 0 2.6099999999999994-1.014285714285716 4.9857142857142875t-2.747142857142858 4.10857142857143-4.107142857142858 2.7457142857142856-4.988571428571429 1.0142857142857125-4.988571428571429-1.0142857142857125-4.107142857142858-2.7457142857142856-2.7471428571428533-4.108571428571434-1.0142857142857142-4.985714285714284 1.0142857142857142-4.990000000000002 2.747142857142857-4.107142857142858 4.107142857142858-2.7457142857142802 4.988571428571429-1.014285714285716q4.53142857142857 0 8.014285714285712 2.8114285714285714l8.525714285714287-8.528571428571428h-5.825714285714284q-0.31428571428571317 0-0.5142857142857125-0.20000000000000018t-0.1999999999999993-0.5142857142857142v-1.4285714285714288q0-0.3114285714285714 0.1999999999999993-0.5114285714285716t0.5142857142857125-0.19999999999999973h9.285714285714285z m-20 31.428571428571427q4.12857142857143 0 7.064285714285717-2.935714285714287t2.9357142857142833-7.064285714285713-2.935714285714287-7.064285714285717-7.064285714285713-2.9357142857142815-7.064285714285715 2.935714285714285-2.935714285714286 7.064285714285713 2.935714285714286 7.064285714285717 7.064285714285715 2.9357142857142833z' })
                )
            );
        }
    }]);

    return FaMars;
}(React.Component);

exports.default = FaMars;
module.exports = exports['default'];