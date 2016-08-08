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

var MdRoom = function (_React$Component) {
    _inherits(MdRoom, _React$Component);

    function MdRoom() {
        _classCallCheck(this, MdRoom);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdRoom).apply(this, arguments));
    }

    _createClass(MdRoom, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm20 19.14q1.7166666666666686 0 2.9299999999999997-1.211666666666666t1.211666666666666-2.9283333333333346-1.211666666666666-2.9333333333333336-2.9299999999999997-1.208333333333334-2.9299999999999997 1.211666666666666-1.211666666666666 2.9300000000000015 1.211666666666666 2.9283333333333346 2.9299999999999997 1.211666666666666z m0-15.780000000000001q4.843333333333334 0 8.241666666666667 3.4000000000000004t3.3999999999999986 8.24q0 2.421666666666667-1.2133333333333347 5.546666666666667t-2.928333333333331 5.859999999999999-3.3999999999999986 5.116666666666667-2.8500000000000014 3.788333333333334l-1.25 1.3283333333333331q-0.46999999999999886-0.5466666666666669-1.25-1.4450000000000003t-2.8166666666666647-3.594999999999999-3.5500000000000025-5.233333333333334-2.7733333333333334-5.741666666666667-1.253333333333332-5.625q0-4.843333333333334 3.4000000000000004-8.241666666666667t8.240000000000002-3.4000000000000004z' })
                )
            );
        }
    }]);

    return MdRoom;
}(React.Component);

exports.default = MdRoom;
module.exports = exports['default'];