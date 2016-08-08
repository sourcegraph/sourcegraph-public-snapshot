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

var MdDirections = function (_React$Component) {
    _inherits(MdDirections, _React$Component);

    function MdDirections() {
        _classCallCheck(this, MdDirections);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdDirections).apply(this, arguments));
    }

    _createClass(MdDirections, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm23.36 24.14l5.783333333333335-5.783333333333335-5.783333333333335-5.858333333333334v4.138333333333332h-8.36q-0.7033333333333331 0-1.1716666666666669 0.5083333333333329t-0.4666666666666668 1.211666666666666v6.643333333333338h3.2799999999999994v-5h6.716666666666669v4.138333333333335z m12.811666666666667-5.311666666666667q1.0933333333333337 1.25 0 2.3433333333333337l-15 15q-0.466666666666665 0.46666666666666856-1.1716666666666669 0.46666666666666856t-1.1716666666666669-0.46666666666666856l-15-15q-0.4666666666666668-0.466666666666665-0.4666666666666668-1.1716666666666669t0.4666666666666668-1.1716666666666669l15-15q0.466666666666665-0.4666666666666668 1.1716666666666669-0.4666666666666668t1.1716666666666669 0.4666666666666668z' })
                )
            );
        }
    }]);

    return MdDirections;
}(React.Component);

exports.default = MdDirections;
module.exports = exports['default'];