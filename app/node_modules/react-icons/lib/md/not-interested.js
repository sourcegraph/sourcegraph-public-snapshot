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

var MdNotInterested = function (_React$Component) {
    _inherits(MdNotInterested, _React$Component);

    function MdNotInterested() {
        _classCallCheck(this, MdNotInterested);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdNotInterested).apply(this, arguments));
    }

    _createClass(MdNotInterested, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm30.546666666666667 28.203333333333337q2.8133333333333326-3.5166666666666657 2.8133333333333326-8.203333333333333 0-5.466666666666667-3.9450000000000003-9.413333333333334t-9.415-3.945000000000003q-4.686666666666667 0-8.200000000000001 2.8133333333333335z m-10.546666666666667 5.156666666666663q4.688333333333333 0 8.203333333333333-2.8133333333333326l-18.75-18.75q-2.8133333333333326 3.5166666666666675-2.8133333333333326 8.203333333333333 0 5.466666666666669 3.9450000000000003 9.413333333333334t9.415 3.9450000000000003z m0-30q6.875 0 11.758333333333333 4.883333333333333t4.883333333333333 11.756666666666668-4.883333333333333 11.759999999999998-11.758333333333333 4.88333333333334-11.758333333333333-4.883333333333333-4.883333333333333-11.760000000000005 4.883333333333333-11.756666666666668 11.758333333333333-4.883333333333332z' })
                )
            );
        }
    }]);

    return MdNotInterested;
}(React.Component);

exports.default = MdNotInterested;
module.exports = exports['default'];