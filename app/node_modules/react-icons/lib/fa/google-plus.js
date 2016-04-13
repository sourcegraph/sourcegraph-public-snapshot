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

var FaGooglePlus = function (_React$Component) {
    _inherits(FaGooglePlus, _React$Component);

    function FaGooglePlus() {
        _classCallCheck(this, FaGooglePlus);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaGooglePlus).apply(this, arguments));
    }

    _createClass(FaGooglePlus, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm24.947777777777777 20.295555555555556q0 3.6111111111111107-1.5111111111111128 6.433333333333334t-4.304444444444446 4.408888888888889-6.407777777777778 1.5888888888888886q-2.586666666666666 0-4.9477777777777785-1.0066666666666677t-4.0611111111111065-2.708888888888886-2.7088888888888887-4.062222222222221-1.0066666666666668-4.948888888888892 1.0066666666666668-4.944444444444443 2.708888888888889-4.062222222222221 4.062222222222223-2.711111111111112 4.9477777777777785-1.0066666666666668q4.966666666666667 0 8.524444444444443 3.333333333333333l-3.4555555555555557 3.315555555555555q-2.030000000000001-1.9622222222222216-5.066666666666666-1.9622222222222216-2.1366666666666667 0-3.9511111111111106 1.0777777777777775t-2.8744444444444444 2.924444444444445-1.057777777777778 4.035555555555554 1.0588888888888892 4.037777777777777 2.8744444444444444 2.925555555555558 3.948888888888888 1.0777777777777793q1.4411111111111108 0 2.647777777777778-0.3999999999999986t1.988888888888889-0.9977777777777774 1.362222222222222-1.3644444444444446 0.8511111111111127-1.4411111111111126 0.37222222222222356-1.2844444444444463h-7.222222222222221v-4.37777777777778h12.014444444444447q0.20777777777777828 1.094444444444445 0.20777777777777828 2.1188888888888897z m15.052222222222223-2.1177777777777784v3.6444444444444457h-3.6288888888888877v3.6288888888888877h-3.6444444444444457v-3.6288888888888877h-3.629999999999999v-3.6444444444444457h3.6288888888888913v-3.6311111111111085h3.6444444444444457v3.6288888888888895h3.6299999999999955z' })
                )
            );
        }
    }]);

    return FaGooglePlus;
}(React.Component);

exports.default = FaGooglePlus;
module.exports = exports['default'];