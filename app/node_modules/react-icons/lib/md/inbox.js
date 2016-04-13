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

var MdInbox = function (_React$Component) {
    _inherits(MdInbox, _React$Component);

    function MdInbox() {
        _classCallCheck(this, MdInbox);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdInbox).apply(this, arguments));
    }

    _createClass(MdInbox, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm31.640000000000004 25v-16.64h-23.360000000000003v16.64h6.716666666666669q0 2.0333333333333314 1.4866666666666681 3.5166666666666657t3.516666666666662 1.4833333333333343 3.5133333333333354-1.4833333333333343 1.4833333333333343-3.5166666666666657h6.641666666666666z m0-20q1.3283333333333367 0 2.34333333333333 0.9766666666666666t1.0166666666666657 2.383333333333333v23.283333333333335q0 1.3266666666666644-1.0166666666666657 2.341666666666665t-2.3433333333333337 1.0166666666666657h-23.36q-1.4066666666666663 0-2.3433333333333337-0.9766666666666666t-0.9366666666666674-2.3849999999999945v-23.28333333333334q0-1.4049999999999985 0.9366666666666665-2.383333333333331t2.3433333333333346-0.9733333333333345h23.36z' })
                )
            );
        }
    }]);

    return MdInbox;
}(React.Component);

exports.default = MdInbox;
module.exports = exports['default'];