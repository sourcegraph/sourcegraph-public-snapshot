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

var MdHdrOn = function (_React$Component) {
    _inherits(MdHdrOn, _React$Component);

    function MdHdrOn() {
        _classCallCheck(this, MdHdrOn);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdHdrOn).apply(this, arguments));
    }

    _createClass(MdHdrOn, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm21.64 22.5v-5h-3.2833333333333314v5h3.2833333333333314z m0-7.5q1.0166666666666657 0 1.7583333333333329 0.7416666666666671t0.7399999999999984 1.7583333333333329v5q0 1.0166666666666657-0.7416666666666671 1.7583333333333329t-1.7583333333333293 0.7416666666666671h-5.783333333333335v-10h5.783333333333335z m-10.780000000000001 3.3599999999999994v-3.3599999999999994h2.5v10h-2.5v-4.140000000000001h-3.3599999999999994v4.140000000000001h-2.5v-10h2.5v3.3599999999999994h3.3599999999999994z m21.64 0.7800000000000011v-1.6400000000000006h-3.3599999999999994v1.6400000000000006h3.3599999999999994z m2.5 0q0 1.5633333333333326-1.4833333333333343 2.3433333333333337l1.4833333333333343 3.5166666666666657h-2.5l-1.4833333333333343-3.361666666666668h-1.8766666666666652v3.361666666666668h-2.5v-10h5.859999999999999q1.0166666666666657 0 1.7583333333333329 0.7400000000000002t0.7416666666666671 1.7599999999999998v1.6383333333333319z' })
                )
            );
        }
    }]);

    return MdHdrOn;
}(React.Component);

exports.default = MdHdrOn;
module.exports = exports['default'];