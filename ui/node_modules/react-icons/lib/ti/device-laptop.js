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

var TiDeviceLaptop = function (_React$Component) {
    _inherits(TiDeviceLaptop, _React$Component);

    function TiDeviceLaptop() {
        _classCallCheck(this, TiDeviceLaptop);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiDeviceLaptop).apply(this, arguments));
    }

    _createClass(TiDeviceLaptop, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm36.65 26.750000000000004c0.013333333333335418-0.5266666666666673 0.01666666666666572-1.0966666666666676 0.01666666666666572-1.75v-15.000000000000004c0-2.756666666666667-2.2433333333333323-5-5-5h-23.33333333333333c-2.7566666666666686 0-5.000000000000002 2.243333333333334-5.000000000000002 5v15c-4.440892098500626e-16 0.6416666666666657 0.0033333333333329662 1.2166666666666686 0.019999999999999574 1.75-1.9083333333333334 0.3766666666666687-3.3533333333333335 2.0666666666666664-3.3533333333333335 4.083333333333336 0 2.296666666666667 1.8700000000000003 4.166666666666664 4.166666666666667 4.166666666666664h31.666666666666668c2.296666666666667 0 4.166666666666664-1.8699999999999974 4.166666666666664-4.166666666666668 0-2.0166666666666657-1.4433333333333351-3.703333333333333-3.3500000000000014-4.083333333333332z m-29.98333333333333-16.750000000000004c-8.881784197001252e-16-0.9166666666666661 0.7499999999999991-1.666666666666666 1.666666666666666-1.666666666666666h23.333333333333336c0.9166666666666643 0 1.6666666666666643 0.75 1.6666666666666643 1.666666666666666v15c0 0.6466666666666683-0.00833333333333286 1.2100000000000009-0.023333333333333428 1.6666666666666679h-1.6433333333333344v-15c0-0.9166666666666661-0.75-1.666666666666666-1.6666666666666679-1.666666666666666h-20c-0.9166666666666661 0-1.666666666666666 0.75-1.666666666666666 1.666666666666666v15h-1.6333333333333329c-0.019999999999999574-0.4400000000000013-0.033333333333333215-1-0.033333333333333215-1.6666666666666679v-15z m23.333333333333336 16.666666666666668h-20.000000000000004v-15h20v15z m5.833333333333332 5h-31.666666666666668c-0.4500000000000002 0-0.8333333333333335-0.38333333333333286-0.8333333333333335-0.8333333333333321s0.3833333333333333-0.8333333333333321 0.8333333333333335-0.8333333333333321h31.666666666666668c0.45000000000000284 0 0.8333333333333357 0.38333333333333286 0.8333333333333357 0.8333333333333321s-0.38333333333333286 0.8333333333333321-0.8333333333333357 0.8333333333333321z' })
                )
            );
        }
    }]);

    return TiDeviceLaptop;
}(React.Component);

exports.default = TiDeviceLaptop;
module.exports = exports['default'];