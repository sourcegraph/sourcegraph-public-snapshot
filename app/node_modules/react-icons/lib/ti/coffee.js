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

var TiCoffee = function (_React$Component) {
    _inherits(TiCoffee, _React$Component);

    function TiCoffee() {
        _classCallCheck(this, TiCoffee);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiCoffee).apply(this, arguments));
    }

    _createClass(TiCoffee, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm28.333333333333336 31.666666666666668h-20c-0.9216666666666686 0-1.6666666666666687-0.745000000000001-1.6666666666666687-1.6666666666666679s0.7450000000000001-1.6666666666666679 1.666666666666667-1.6666666666666679h20c0.9216666666666669 0 1.6666666666666679 0.745000000000001 1.6666666666666679 1.6666666666666679s-0.745000000000001 1.6666666666666679-1.6666666666666679 1.6666666666666679z m0.8333333333333321-23.333333333333336h-20.833333333333336v15.000000000000004c1.7763568394002505e-15 1.8333333333333321 1.5000000000000018 3.333333333333332 3.3333333333333357 3.333333333333332h13.333333333333332c1.8333333333333321 0 3.333333333333332-1.5 3.333333333333332-3.333333333333332v-3.333333333333332h0.8333333333333321c3.2166666666666686 0 5.833333333333336-2.616666666666667 5.833333333333336-5.833333333333334s-2.616666666666667-5.833333333333336-5.833333333333332-5.833333333333336z m-4.166666666666668 15.000000000000004h-13.333333333333332v-11.666666666666668h13.333333333333332v11.666666666666668z m4.166666666666668-6.666666666666668h-2.5v-5h2.5c1.3783333333333339 0 2.5 1.1216666666666661 2.5 2.5s-1.1216666666666661 2.5-2.5 2.5z' })
                )
            );
        }
    }]);

    return TiCoffee;
}(React.Component);

exports.default = TiCoffee;
module.exports = exports['default'];