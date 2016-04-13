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

var FaCloudUpload = function (_React$Component) {
    _inherits(FaCloudUpload, _React$Component);

    function FaCloudUpload() {
        _classCallCheck(this, FaCloudUpload);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaCloudUpload).apply(this, arguments));
    }

    _createClass(FaCloudUpload, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm26.666666666666664 19.333333333333332q0-0.293333333333333-0.18666666666666742-0.4800000000000004l-7.333333333333332-7.333333333333332q-0.1893333333333338-0.18666666666666742-0.4800000000000004-0.18666666666666742t-0.4800000000000004 0.18666666666666742l-7.310666666666666 7.313333333333333q-0.2093333333333316 0.24933333333333252-0.2093333333333316 0.5 0 0.293333333333333 0.1893333333333338 0.4800000000000004t0.4800000000000004 0.18666666666666742h4.664v7.333333333333332q0 0.27066666666666706 0.1960000000000015 0.46933333333333493t0.466666666666665 0.19733333333333292h4q0.27199999999999847 0 0.47066666666666634-0.19733333333333292t0.1999999999999993-0.46933333333333493v-7.333333333333332h4.666666666666668q0.27199999999999847 0 0.47066666666666634-0.19733333333333292t0.19599999999999795-0.46933333333333493z m13.333333333333336 6q0 3.312000000000001-2.344000000000001 5.655999999999999t-5.655999999999999 2.3439999999999976h-22.666666666666668q-3.8533333333333335 0-6.593333333333334-2.740000000000002t-2.7399999999999984-6.593333333333327q0-2.706666666666667 1.4586666666666668-5t3.9173333333333336-3.437333333333333q-0.040000000000000036-0.6266666666666669-0.040000000000000036-0.8960000000000008 0-4.417333333333334 3.1240000000000006-7.541333333333332t7.539999999999999-3.1253333333333337q3.253333333333334 0 5.949333333333332 1.8133333333333335t3.926666666666666 4.810666666666665q1.4800000000000004-1.293333333333333 3.458666666666666-1.293333333333333 2.2093333333333334 0 3.7733333333333334 1.564t1.5586666666666673 3.7720000000000002q0 1.5826666666666664-0.8533333333333317 2.873333333333333 2.7066666666666634 0.6453333333333333 4.448 2.8226666666666667t1.738666666666667 4.970666666666666z' })
                )
            );
        }
    }]);

    return FaCloudUpload;
}(React.Component);

exports.default = FaCloudUpload;
module.exports = exports['default'];