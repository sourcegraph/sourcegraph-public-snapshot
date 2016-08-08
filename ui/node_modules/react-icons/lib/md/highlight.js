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

var MdHighlight = function (_React$Component) {
    _inherits(MdHighlight, _React$Component);

    function MdHighlight() {
        _classCallCheck(this, MdHighlight);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdHighlight).apply(this, arguments));
    }

    _createClass(MdHighlight, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm28.28333333333333 10.938333333333333l3.5133333333333354-3.5166666666666666 2.3433333333333337 2.3449999999999998-3.5166666666666657 3.5933333333333337z m-22.424999999999997-1.171666666666665l2.343333333333332-2.3450000000000006 3.5166666666666675 3.5166666666666657-2.3450000000000006 2.42z m12.5-6.406666666666666h3.2833333333333314v5h-3.2833333333333314v-5z m-8.358333333333334 20v-8.36h20v8.36l-5 5v8.283333333333331h-10v-8.283333333333331z' })
                )
            );
        }
    }]);

    return MdHighlight;
}(React.Component);

exports.default = MdHighlight;
module.exports = exports['default'];