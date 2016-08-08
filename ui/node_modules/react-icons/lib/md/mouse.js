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

var MdMouse = function (_React$Component) {
    _inherits(MdMouse, _React$Component);

    function MdMouse() {
        _classCallCheck(this, MdMouse);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdMouse).apply(this, arguments));
    }

    _createClass(MdMouse, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm18.36 1.7966666666666669v13.203333333333333h-11.716666666666667q0-5.078333333333333 3.3966666666666665-8.828333333333333t8.32-4.375z m-11.719999999999999 23.203333333333333v-6.640000000000001h26.71666666666667v6.640000000000001q0 5.466666666666669-3.943333333333335 9.413333333333334t-9.413333333333334 3.9450000000000003-9.413333333333334-3.9450000000000003-3.9449999999999994-9.413333333333334z m15-23.203333333333337q4.921666666666667 0.625 8.32 4.375t3.3999999999999986 8.828333333333337h-11.719999999999999v-13.203333333333333z' })
                )
            );
        }
    }]);

    return MdMouse;
}(React.Component);

exports.default = MdMouse;
module.exports = exports['default'];