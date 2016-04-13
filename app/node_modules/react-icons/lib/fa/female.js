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

var FaFemale = function (_React$Component) {
    _inherits(FaFemale, _React$Component);

    function FaFemale() {
        _classCallCheck(this, FaFemale);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaFemale).apply(this, arguments));
    }

    _createClass(FaFemale, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm34.285714285714285 23.571428571428573q0 0.8928571428571423-0.6257142857142881 1.5171428571428578t-1.5171428571428507 0.6257142857142846q-1.138571428571428 0-1.7857142857142847-0.9600000000000009l-5.067142857142862-7.611428571428572h-1.0042857142857144v2.9471428571428575l5.5142857142857125 9.174285714285713q0.1999999999999993 0.33428571428571274 0.1999999999999993 0.7371428571428567 0 0.581428571428571-0.4242857142857126 1.0057142857142871t-1.004285714285711 0.4214285714285744h-4.285714285714285v6.071428571428569q0 1.028571428571425-0.7371428571428567 1.7642857142857125t-1.762857142857147 0.7357142857142875h-3.571428571428573q-1.0285714285714285 0-1.7628571428571433-0.7357142857142875t-0.7371428571428531-1.7642857142857125v-6.071428571428569h-4.2857142857142865q-0.5800000000000001 0-1.0042857142857144-0.4228571428571435t-0.4242857142857144-1.0057142857142871q0-0.3999999999999986 0.1999999999999993-0.7357142857142875l5.514285714285716-9.174285714285713v-2.9471428571428575h-1.0042857142857144l-5.0671428571428585 7.611428571428572q-0.647142857142855 0.9600000000000009-1.7857142857142847 0.9600000000000009-0.8928571428571432 0-1.5171428571428578-0.6257142857142846t-0.6257142857142854-1.5171428571428578q0-0.6471428571428568 0.35714285714285676-1.1828571428571415l5.7142857142857135-8.571428571428571q1.62857142857143-2.3885714285714315 3.9285714285714306-2.3885714285714315h8.57142857142857q2.3000000000000007 0 3.9285714285714306 2.388571428571428l5.714285714285715 8.571428571428571q0.3571428571428541 0.5357142857142847 0.3571428571428541 1.1828571428571415z m-9.285714285714285-17.857142857142858q0 2.0757142857142856-1.46142857142857 3.5385714285714265t-3.53857142857143 1.4614285714285735-3.53857142857143-1.4614285714285717-1.46142857142857-3.538571428571429 1.46142857142857-3.5385714285714283 3.53857142857143-1.4614285714285717 3.53857142857143 1.4614285714285713 1.46142857142857 3.5385714285714287z' })
                )
            );
        }
    }]);

    return FaFemale;
}(React.Component);

exports.default = FaFemale;
module.exports = exports['default'];