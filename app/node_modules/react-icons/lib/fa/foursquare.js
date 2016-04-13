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

var FaFoursquare = function (_React$Component) {
    _inherits(FaFoursquare, _React$Component);

    function FaFoursquare() {
        _classCallCheck(this, FaFoursquare);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaFoursquare).apply(this, arguments));
    }

    _createClass(FaFoursquare, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm28.035714285714285 9.685714285714287l0.8257142857142838-4.328571428571428q0.1114285714285721-0.5142857142857142-0.1999999999999993-0.8928571428571432t-0.7828571428571429-0.3799999999999999h-15.892857142857139q-0.5142857142857142 0-0.8599999999999994 0.378571428571429t-0.3457142857142852 0.8257142857142856v24.575714285714284q0 0.15714285714285836 0.13428571428571345 0.022857142857141355l6.495714285714287-7.857142857142858q0.5142857142857125-0.5800000000000018 0.8485714285714288-0.7471428571428582t1.071428571428573-0.16857142857142904h5.335714285714285q0.4914285714285711 0 0.8257142857142874-0.32428571428571473t0.3999999999999986-0.6571428571428584q0.5371428571428574-2.902857142857144 0.8285714285714292-4.264285714285714 0.08857142857142719-0.468571428571428-0.25714285714285623-0.8928571428571423t-0.8142857142857132-0.4242857142857144h-6.567142857142866q-0.6471428571428568 0-1.071428571428573-0.4242857142857144t-0.4242857142857126-1.0714285714285712v-0.935714285714285q0-0.6471428571428568 0.42571428571428527-1.0600000000000005t1.071428571428573-0.4142857142857146h7.722857142857144q0.3999999999999986 0 0.7814285714285703-0.3000000000000007t0.4485714285714302-0.6571428571428566z m5.067142857142855-4.952857142857143q-0.33428571428571274 1.6285714285714281-1.1942857142857157 5.948571428571428t-1.5514285714285698 7.814285714285713-0.7814285714285703 3.87142857142857q-0.13428571428571345 0.4914285714285711-0.1999999999999993 0.725714285714286t-0.31428571428571317 0.725714285714286-0.5471428571428589 0.7371428571428567-0.8599999999999994 0.4671428571428571-1.2942857142857136 0.2228571428571442h-6.048571428571428q-0.28999999999999915 0-0.4914285714285711 0.2228571428571442-0.17857142857142705 0.1999999999999993-9.508571428571429 11.025714285714287-0.4914285714285711 0.5571428571428569-1.305714285714286 0.6357142857142861t-1.0828571428571427-0.12285714285714278q-1.2285714285714286-0.4914285714285711-1.2285714285714286-2.1857142857142833v-31.472857142857148q0-1.228571428571429 0.8499999999999996-2.288571428571429t2.6800000000000015-1.0599999999999996h19.81857142857143q2.1214285714285737 0 2.835714285714289 1.1857142857142857t0.2228571428571442 3.547142857142857z m0 0l-3.5285714285714285 17.634285714285717q0.09142857142857252-0.379999999999999 0.7828571428571429-3.87142857142857t1.5514285714285734-7.814285714285713 1.1942857142857122-5.948571428571428z' })
                )
            );
        }
    }]);

    return FaFoursquare;
}(React.Component);

exports.default = FaFoursquare;
module.exports = exports['default'];