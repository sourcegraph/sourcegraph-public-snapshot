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

var MdNature = function (_React$Component) {
    _inherits(MdNature, _React$Component);

    function MdNature() {
        _classCallCheck(this, MdNature);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdNature).apply(this, arguments));
    }

    _createClass(MdNature, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm21.64 26.875v6.483333333333334h10v3.2833333333333314h-23.28333333333333v-3.2833333333333314h10v-6.558333333333337q-4.216666666666667-0.7033333333333331-6.99-3.9450000000000003t-2.773333333333335-7.53833333333333q0-4.843333333333334 3.4383333333333344-8.283333333333333t8.28-3.4366666666666665 8.241666666666667 3.4383333333333335 3.3999999999999986 8.280000000000001q0 4.453333333333333-2.969999999999999 7.733333333333334t-7.341666666666669 3.830000000000002z' })
                )
            );
        }
    }]);

    return MdNature;
}(React.Component);

exports.default = MdNature;
module.exports = exports['default'];