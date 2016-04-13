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

var MdEvent = function (_React$Component) {
    _inherits(MdEvent, _React$Component);

    function MdEvent() {
        _classCallCheck(this, MdEvent);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdEvent).apply(this, arguments));
    }

    _createClass(MdEvent, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm31.640000000000004 31.640000000000004v-18.28333333333334h-23.28333333333334v18.283333333333335h23.283333333333335z m-5-30h3.359999999999996v3.359999999999996h1.6400000000000006q1.3283333333333331 0 2.3433333333333337 1.0166666666666666t1.0166666666666657 2.3400000000000007v23.28333333333333q0 1.326666666666668-1.0166666666666657 2.3416666666666686t-2.3433333333333337 1.0166666666666657h-23.28333333333333q-1.405000000000002 0-2.3833333333333346-1.0166666666666657t-0.9733333333333345-2.341666666666665v-23.28333333333334q0-1.3266666666666653 0.9749999999999996-2.341666666666665t2.383333333333333-1.0150000000000006h1.6416666666666675v-3.36h3.3583333333333343v3.36h13.283333333333331v-3.36z m1.7199999999999953 18.359999999999996v8.36h-8.36v-8.36h8.36z' })
                )
            );
        }
    }]);

    return MdEvent;
}(React.Component);

exports.default = MdEvent;
module.exports = exports['default'];