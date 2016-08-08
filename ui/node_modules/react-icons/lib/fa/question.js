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

var FaQuestion = function (_React$Component) {
    _inherits(FaQuestion, _React$Component);

    function FaQuestion() {
        _classCallCheck(this, FaQuestion);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaQuestion).apply(this, arguments));
    }

    _createClass(FaQuestion, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm23.571428571428573 28.035714285714285v5.357142857142854q0 0.3571428571428541-0.2671428571428578 0.6257142857142881t-0.6257142857142846 0.2671428571428578h-5.357142857142858q-0.35714285714285765 0-0.6257142857142846-0.2671428571428578t-0.2671428571428578-0.625714285714281v-5.357142857142858q0-0.35714285714285765 0.2671428571428578-0.6257142857142846t0.6257142857142846-0.26714285714286135h5.357142857142858q0.35714285714285765 0 0.6257142857142846 0.2671428571428578t0.2671428571428578 0.6257142857142846z m7.0528571428571425-13.392857142857142q0 1.2057142857142864-0.345714285714287 2.2542857142857144t-0.7814285714285703 1.707142857142859-1.2285714285714278 1.3285714285714292-1.2814285714285703 0.9714285714285715-1.3599999999999994 0.7928571428571445q-0.9142857142857146 0.514285714285716-1.5285714285714285 1.451428571428572t-0.6142857142857139 1.4957142857142856q0 0.38142857142857167-0.2671428571428578 0.7285714285714278t-0.6242857142857154 0.34285714285714164h-5.357142857142858q-0.33428571428571274 0-0.5685714285714276-0.4114285714285728t-0.23857142857142932-0.8385714285714272v-1.0042857142857144q0-1.8528571428571432 1.452857142857141-3.4928571428571438t3.1914285714285704-2.421428571428571q1.3185714285714276-0.6028571428571432 1.8771428571428572-1.25t0.5571428571428569-1.6971428571428575q0-0.9371428571428577-1.0371428571428574-1.6514285714285712t-2.3999999999999986-0.7142857142857135q-1.451428571428572 0-2.41 0.6471428571428568-0.7800000000000011 0.5571428571428569-2.385714285714286 2.5671428571428567-0.2914285714285718 0.35714285714285765-0.6928571428571431 0.35714285714285765-0.2671428571428578 0-0.5571428571428569-0.17857142857142883l-3.6671428571428564-2.7914285714285736q-0.28999999999999915-0.22285714285714242-0.3457142857142852-0.5571428571428569t0.12285714285714278-0.6285714285714281q3.571428571428571-5.935714285714286 10.357142857142856-5.935714285714286 1.7857142857142847 0 3.5942857142857143 0.6928571428571431t3.2571428571428562 1.8528571428571423 2.3714285714285737 2.845714285714287 0.9142857142857146 3.5385714285714283z' })
                )
            );
        }
    }]);

    return FaQuestion;
}(React.Component);

exports.default = FaQuestion;
module.exports = exports['default'];