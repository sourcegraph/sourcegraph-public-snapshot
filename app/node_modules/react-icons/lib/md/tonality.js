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

var MdTonality = function (_React$Component) {
    _inherits(MdTonality, _React$Component);

    function MdTonality() {
        _classCallCheck(this, MdTonality);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdTonality).apply(this, arguments));
    }

    _createClass(MdTonality, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm32.89000000000001 23.36q0.1566666666666663-0.466666666666665 0.3133333333333326-1.7166666666666686h-11.563333333333333v1.7166666666666686h11.25z m-2.5 5q0.9383333333333326-1.3283333333333331 1.1716666666666669-1.7166666666666686h-9.921666666666667v1.7166666666666686h8.75z m-8.75 4.843333333333334q2.5-0.3133333333333326 4.843333333333334-1.5633333333333326h-4.843333333333334v1.5633333333333326z m0-16.563333333333336v1.716666666666665h11.563333333333333q-0.1566666666666663-1.25-0.3133333333333326-1.7166666666666686h-11.25z m0-5v1.7166666666666668h9.921666666666667q-0.23333333333333428-0.38833333333333364-1.1716666666666669-1.7166666666666668h-8.75z m0-4.843333333333334v1.5633333333333326h4.843333333333334q-2.3433333333333337-1.25-4.843333333333334-1.5633333333333335z m-3.280000000000001 26.40666666666667v-26.406666666666666q-4.921666666666667 0.6250000000000009-8.32 4.375t-3.4000000000000057 8.828333333333333 3.4000000000000004 8.828333333333333 8.319999999999999 4.375z m1.6399999999999935-29.843333333333334q6.875 8.881784197001252e-16 11.758333333333333 4.883333333333335t4.883333333333333 11.756666666666666-4.883333333333333 11.759999999999998-11.758333333333333 4.88333333333334-11.758333333333333-4.883333333333333-4.883333333333333-11.760000000000005 4.883333333333333-11.756666666666668 11.758333333333333-4.883333333333332z' })
                )
            );
        }
    }]);

    return MdTonality;
}(React.Component);

exports.default = MdTonality;
module.exports = exports['default'];