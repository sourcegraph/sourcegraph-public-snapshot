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

var MdWifiTethering = function (_React$Component) {
    _inherits(MdWifiTethering, _React$Component);

    function MdWifiTethering() {
        _classCallCheck(this, MdWifiTethering);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdWifiTethering).apply(this, arguments));
    }

    _createClass(MdWifiTethering, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm20 5q6.875 0 11.758333333333333 4.843333333333334t4.883333333333333 11.796666666666667q0 4.533333333333335-2.2266666666666666 8.399999999999999t-6.056666666666665 6.053333333333335l-1.7166666666666686-2.8900000000000006q3.0450000000000017-1.7966666666666669 4.883333333333333-4.883333333333333t1.8333333333333357-6.68q0-5.466666666666669-3.905000000000001-9.375t-9.453333333333333-3.908333333333333-9.45 3.91-3.908333333333333 9.373333333333333q0 3.671666666666667 1.7966666666666669 6.716666666666669t4.843333333333334 4.844999999999999l-1.6416666666666675 2.8916666666666657q-3.828333333333333-2.1883333333333326-6.055000000000001-6.055t-2.226666666666667-8.399999999999999q0-6.949999999999999 4.883333333333335-11.795t11.758333333333333-4.8433333333333355z m10 16.64q0 2.7333333333333343-1.3666666666666671 5.038333333333334t-3.633333333333333 3.6333333333333364l-1.6400000000000006-2.8900000000000006q3.2833333333333314-1.9533333333333331 3.2833333333333314-5.783333333333335 0-2.7333333333333343-1.9549999999999983-4.686666666666667t-4.688333333333333-1.951666666666668-4.688333333333333 1.9499999999999993-1.9533333333333331 4.690000000000001q0 3.828333333333333 3.2833333333333314 5.783333333333335l-1.6416666666666657 2.8866666666666667q-2.2666666666666675-1.3283333333333331-3.633333333333333-3.633333333333333t-1.3666666666666671-5.036666666666669q0-4.140000000000001 2.9299999999999997-7.0699999999999985t7.07-2.9300000000000015 7.07 2.9299999999999997 2.9299999999999997 7.07z m-10-3.280000000000001q1.3283333333333331 0 2.3433333333333337 0.9766666666666666t1.0166666666666657 2.3049999999999997-1.0166666666666657 2.3433333333333337-2.3433333333333337 1.0150000000000006-2.3433333333333337-1.0166666666666657-1.0166666666666657-2.3416666666666686 1.0166666666666657-2.3049999999999997 2.3433333333333337-0.9766666666666666z' })
                )
            );
        }
    }]);

    return MdWifiTethering;
}(React.Component);

exports.default = MdWifiTethering;
module.exports = exports['default'];