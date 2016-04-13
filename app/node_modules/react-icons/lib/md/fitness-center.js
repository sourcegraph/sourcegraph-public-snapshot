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

var MdFitnessCenter = function (_React$Component) {
    _inherits(MdFitnessCenter, _React$Component);

    function MdFitnessCenter() {
        _classCallCheck(this, MdFitnessCenter);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdFitnessCenter).apply(this, arguments));
    }

    _createClass(MdFitnessCenter, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm34.29666666666667 24.766666666666666l2.3433333333333337 2.4200000000000017-3.5166666666666657 3.5166666666666657 2.344999999999999 2.4200000000000017-2.3416666666666686 2.3433333333333337-2.421666666666667-2.3433333333333337-3.5166666666666657 3.5166666666666657-2.4200000000000017-2.344999999999999-2.3433333333333337 2.3433333333333337-2.424999999999997-2.3433333333333337 5.940000000000001-5.938333333333333-14.296666666666667-14.296666666666667-5.938333333333334 5.939999999999998-2.3433333333333333-2.423333333333332 2.3433333333333333-2.3433333333333337-2.3433333333333333-2.421666666666667 3.516666666666667-3.5166666666666657-2.3449999999999998-2.418333333333334 2.3433333333333337-2.3433333333333337 2.423333333333333 2.339999999999999 3.5133333333333336-3.5166666666666666 2.42 2.3483333333333336 2.344999999999999-2.3433333333333333 2.421666666666667 2.3416666666666672-5.938333333333334 5.9383333333333335 14.296666666666669 14.296666666666665 5.938333333333333-5.938333333333333 2.3433333333333337 2.421666666666667z' })
                )
            );
        }
    }]);

    return MdFitnessCenter;
}(React.Component);

exports.default = MdFitnessCenter;
module.exports = exports['default'];