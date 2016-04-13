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

var MdFiberSmartRecord = function (_React$Component) {
    _inherits(MdFiberSmartRecord, _React$Component);

    function MdFiberSmartRecord() {
        _classCallCheck(this, MdFiberSmartRecord);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdFiberSmartRecord).apply(this, arguments));
    }

    _createClass(MdFiberSmartRecord, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm28.36 7.11q4.375 1.0933333333333328 7.188333333333333 4.6883333333333335t2.8116666666666674 8.201666666666666-2.8133333333333326 8.204999999999998-7.190000000000001 4.688333333333333v-3.4383333333333326q2.966666666666665-1.0166666666666657 4.805000000000003-3.5933333333333337t1.8383333333333312-5.861666666666665-1.8333333333333357-5.858333333333334-4.806666666666665-3.5933333333333337v-3.4399999999999986z m-26.72 12.89q-4.440892098500626e-16-5.466666666666667 3.9450000000000003-9.413333333333334t9.415-3.9449999999999994 9.411666666666669 3.9449999999999994 3.9449999999999967 9.413333333333334-3.9450000000000003 9.413333333333334-9.411666666666665 3.9450000000000003-9.416666666666668-3.9450000000000003-3.9416666666666655-9.413333333333334z' })
                )
            );
        }
    }]);

    return MdFiberSmartRecord;
}(React.Component);

exports.default = MdFiberSmartRecord;
module.exports = exports['default'];