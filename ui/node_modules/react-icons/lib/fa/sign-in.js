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

var FaSignIn = function (_React$Component) {
    _inherits(FaSignIn, _React$Component);

    function FaSignIn() {
        _classCallCheck(this, FaSignIn);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaSignIn).apply(this, arguments));
    }

    _createClass(FaSignIn, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm29.285714285714285 20q0 0.5799999999999983-0.4242857142857126 1.0042857142857144l-12.142857142857142 12.142857142857139q-0.4242857142857126 0.42428571428571615-1.0042857142857144 0.42428571428571615t-1.0042857142857144-0.42428571428571615-0.4242857142857144-1.0042857142857073v-6.428571428571431h-10q-0.580000000000001 0-1.0042857142857153-0.4242857142857126t-0.42428571428571393-1.004285714285718v-8.571428571428571q0-0.5800000000000001 0.4242857142857144-1.0042857142857144t1.004285714285714-0.4242857142857126h10v-6.428571428571429q0-0.5800000000000001 0.4242857142857144-1.0042857142857144t1.0042857142857162-0.4242857142857144 1.0042857142857144 0.4242857142857144l12.142857142857142 12.142857142857142q0.4242857142857126 0.4242857142857126 0.4242857142857126 1.0042857142857144z m7.857142857142861-7.857142857142858v15.714285714285715q0 2.6571428571428584-1.8857142857142861 4.542857142857141t-4.5428571428571445 1.8857142857142861h-7.142857142857142q-0.28999999999999915 0-0.5028571428571418-0.21142857142856997t-0.21142857142857352-0.5028571428571453q0-0.09000000000000341-0.022857142857141355-0.4471428571428575t-0.011428571428570677-0.5914285714285725 0.0671428571428585-0.5242857142857176 0.2228571428571442-0.43571428571428683 0.4571428571428555-0.14285714285714235h7.142857142857142q1.471428571428575 0 2.5214285714285722-1.0500000000000007t1.0514285714285663-2.5228571428571342v-15.714285714285715q0-1.474285714285715-1.0499999999999972-2.524285714285714t-2.5214285714285722-1.0471428571428572h-6.965714285714284l-0.25714285714285623-0.024285714285714022-0.25714285714285623-0.06714285714285673-0.17857142857142705-0.12285714285714278-0.15714285714285836-0.1999999999999993-0.04285714285714448-0.3028571428571425q0-0.08999999999999986-0.024285714285714022-0.4471428571428575t-0.011428571428570677-0.5914285714285716 0.0671428571428585-0.5228571428571431 0.2228571428571442-0.43571428571428594 0.4571428571428555-0.1471428571428568h7.142857142857142q2.657142857142855 0 4.542857142857141 1.8857142857142861t1.8857142857142861 4.5428571428571445z' })
                )
            );
        }
    }]);

    return FaSignIn;
}(React.Component);

exports.default = FaSignIn;
module.exports = exports['default'];