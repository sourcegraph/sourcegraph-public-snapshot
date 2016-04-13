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

var MdTrackChanges = function (_React$Component) {
    _inherits(MdTrackChanges, _React$Component);

    function MdTrackChanges() {
        _classCallCheck(this, MdTrackChanges);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdTrackChanges).apply(this, arguments));
    }

    _createClass(MdTrackChanges, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm31.796666666666667 8.203333333333333q4.843333333333341 4.843333333333334 4.843333333333341 11.796666666666667 0 6.875-4.883333333333333 11.758333333333333t-11.756666666666675 4.883333333333333-11.76-4.883333333333333-4.883333333333334-11.758333333333333 4.883333333333334-11.758333333333333 11.76-4.883333333333333h1.6383333333333319v13.75q1.716666666666665 0.9383333333333326 1.716666666666665 2.8900000000000006 0 1.326666666666668-1.0133333333333319 2.3416666666666686t-2.3433333333333337 1.0166666666666657-2.344999999999999-1.0166666666666657-1.014999999999997-2.3400000000000034q0-1.9549999999999983 1.7199999999999989-2.8916666666666657v-3.5166666666666657q-2.1883333333333326 0.6266666666666669-3.5933333333333337 2.3450000000000006t-1.408333333333335 4.063333333333331q0 2.7333333333333343 1.9533333333333331 4.688333333333333t4.690000000000001 1.9533333333333331 4.688333333333333-1.9533333333333331 1.9533333333333331-4.688333333333333q0-2.578333333333333-1.9533333333333331-4.688333333333333l2.344999999999999-2.3450000000000006q2.9666666666666686 2.9716666666666676 2.9666666666666686 7.033333333333333 0 4.140000000000001-2.9299999999999997 7.07t-7.07 2.9299999999999997-7.07-2.9299999999999997-2.9299999999999997-7.07q0-3.671666666666667 2.383333333333333-6.445t5.976666666666667-3.4000000000000004v-3.3549999999999986q-4.921666666666667 0.625-8.32 4.375t-3.3999999999999986 8.825q0 5.466666666666669 3.946666666666667 9.413333333333334t9.413333333333332 3.9450000000000003 9.413333333333334-3.9450000000000003 3.9450000000000003-9.413333333333334q0-5.546666666666667-3.9066666666666663-9.453333333333333z' })
                )
            );
        }
    }]);

    return MdTrackChanges;
}(React.Component);

exports.default = MdTrackChanges;
module.exports = exports['default'];