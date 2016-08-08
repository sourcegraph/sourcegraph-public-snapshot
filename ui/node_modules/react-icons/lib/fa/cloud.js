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

var FaCloud = function (_React$Component) {
    _inherits(FaCloud, _React$Component);

    function FaCloud() {
        _classCallCheck(this, FaCloud);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaCloud).apply(this, arguments));
    }

    _createClass(FaCloud, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm40 25.333333333333332q0 3.312000000000001-2.344000000000001 5.655999999999999t-5.655999999999999 2.3439999999999976h-22.666666666666668q-3.8533333333333335 0-6.593333333333334-2.740000000000002t-2.7399999999999984-6.593333333333327q0-2.7506666666666675 1.48-5.030666666666669t3.893333333333333-3.4066666666666663q-0.040000000000000036-0.5839999999999996-0.040000000000000036-0.895999999999999 0-4.417333333333334 3.125333333333333-7.541333333333332t7.541333333333334-3.1253333333333337q3.293333333333333 0 5.969333333333331 1.833333333333333t3.9066666666666663 4.792000000000001q1.4573333333333345-1.2920000000000016 3.4573333333333345-1.2920000000000016 2.2093333333333334 0 3.773333333333337 1.5626666666666669t1.5599999999999952 3.770666666666667q0 1.5626666666666669-0.8533333333333317 2.874666666666668 2.6880000000000024 0.6266666666666652 4.437333333333335 2.802666666666667t1.7493333333333325 4.989333333333331z' })
                )
            );
        }
    }]);

    return FaCloud;
}(React.Component);

exports.default = FaCloud;
module.exports = exports['default'];