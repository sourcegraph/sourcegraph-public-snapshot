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

var MdFilterTiltShift = function (_React$Component) {
    _inherits(MdFilterTiltShift, _React$Component);

    function MdFilterTiltShift() {
        _classCallCheck(this, MdFilterTiltShift);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdFilterTiltShift).apply(this, arguments));
    }

    _createClass(MdFilterTiltShift, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm9.453333333333333 32.89000000000001l2.3433333333333337-2.3433333333333337q2.8900000000000006 2.1883333333333326 6.563333333333333 2.6566666666666663v3.3599999999999994q-4.921666666666667-0.46666666666666856-8.906666666666668-3.671666666666667z m12.186666666666667 0.3133333333333326q3.671666666666667-0.46666666666666856 6.483333333333334-2.6566666666666663l2.423333333333332 2.3433333333333337q-3.9833333333333343 3.203333333333333-8.906666666666666 3.671666666666667v-3.3616666666666717z m8.906666666666666-5q2.1883333333333326-2.9666666666666686 2.6566666666666663-6.483333333333334h3.3599999999999994q-0.46666666666666856 4.841666666666669-3.671666666666667 8.826666666666668z m-5.546666666666667-8.20333333333334q0 2.0333333333333314-1.4833333333333343 3.5166666666666657t-3.5166666666666657 1.4833333333333343-3.5166666666666657-1.4833333333333343-1.4833333333333343-3.5166666666666657 1.4833333333333343-3.5166666666666657 3.5166666666666657-1.4833333333333343 3.5166666666666657 1.4833333333333343 1.4833333333333343 3.5166666666666657z m-18.203333333333337 1.6400000000000006q0.4666666666666668 3.671666666666667 2.6566666666666663 6.483333333333334l-2.3433333333333293 2.4266666666666623q-3.203333333333333-3.9833333333333343-3.6716666666666664-8.906666666666666h3.361666666666667z m2.6566666666666663-9.843333333333334q-2.186666666666662 2.889999999999999-2.656666666666662 6.563333333333333h-3.360000000000001q0.4666666666666668-4.921666666666667 3.671666666666666-8.906666666666668z m23.750000000000004 6.563333333333333q-0.46666666666666856-3.671666666666667-2.6566666666666663-6.563333333333333l2.3433333333333337-2.3433333333333337q3.203333333333333 3.9833333333333343 3.671666666666667 8.906666666666666h-3.3616666666666646z m-2.6566666666666663-11.25l-2.3433333333333337 2.3433333333333337q-2.8900000000000006-2.1883333333333335-6.563333333333333-2.6566666666666663v-3.3600000000000003q4.921666666666667 0.4666666666666668 8.906666666666666 3.671666666666666z m-12.186666666666667-0.31333333333333346q-3.671666666666667 0.4666666666666668-6.563333333333333 2.656666666666667l-2.3433333333333337-2.343333333333333q3.9833333333333343-3.203333333333333 8.906666666666666-3.6716666666666664v3.361666666666667z' })
                )
            );
        }
    }]);

    return MdFilterTiltShift;
}(React.Component);

exports.default = MdFilterTiltShift;
module.exports = exports['default'];