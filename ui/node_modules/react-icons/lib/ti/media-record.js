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

var TiMediaRecord = function (_React$Component) {
    _inherits(TiMediaRecord, _React$Component);

    function TiMediaRecord() {
        _classCallCheck(this, TiMediaRecord);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiMediaRecord).apply(this, arguments));
    }

    _createClass(TiMediaRecord, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm30 20c0-2.7616666666666667-1.120000000000001-5.261666666666667-2.9283333333333346-7.071666666666667-1.8099999999999987-1.8083333333333336-4.309999999999999-2.928333333333333-7.071666666666665-2.928333333333333-2.759999999999998 0-5.260000000000002 1.120000000000001-7.07 2.928333333333333-1.8100000000000005 1.8100000000000005-2.9299999999999997 4.3100000000000005-2.9299999999999997 7.071666666666667 0 2.759999999999998 1.120000000000001 5.260000000000002 2.9299999999999997 7.07s4.309999999999999 2.9299999999999997 7.07 2.9299999999999997c2.7616666666666667 0 5.261666666666667-1.120000000000001 7.071666666666665-2.9299999999999997 1.8083333333333336-1.8099999999999987 2.9283333333333346-4.309999999999999 2.9283333333333346-7.07z' })
                )
            );
        }
    }]);

    return TiMediaRecord;
}(React.Component);

exports.default = TiMediaRecord;
module.exports = exports['default'];