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

var MdLocalPlay = function (_React$Component) {
    _inherits(MdLocalPlay, _React$Component);

    function MdLocalPlay() {
        _classCallCheck(this, MdLocalPlay);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdLocalPlay).apply(this, arguments));
    }

    _createClass(MdLocalPlay, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm25.938333333333336 27.96666666666667l-1.7966666666666669-6.795000000000002 5.466666666666665-4.533333333333335-7.030000000000001-0.38833333333333186-2.578333333333333-6.563333333333333-2.578333333333333 6.563333333333333-7.109999999999999 0.39000000000000057 5.546666666666667 4.533333333333335-1.7966666666666669 6.794999999999998 5.938333333333333-3.826666666666668z m7.421666666666663-7.966666666666669q0 1.3283333333333331 0.9766666666666666 2.3433333333333337t2.3049999999999997 1.0166666666666657v6.638333333333335q0 1.326666666666668-0.9766666666666666 2.3416666666666686t-2.306666666666665 1.0166666666666657h-26.715q-1.330000000000001 0-2.3066666666666675-1.0166666666666657t-0.9766666666666666-2.3400000000000034v-6.641666666666666q1.4066666666666672 0 2.343333333333333-0.9766666666666666t0.9383333333333335-2.3816666666666677q0-1.3299999999999983-0.9766666666666666-2.344999999999999t-2.3066666666666666-1.0166666666666657v-6.638333333333335q0-1.3283333333333331 0.9766666666666666-2.3433333333333337t2.3050000000000006-1.0166666666666666h26.71666666666667q1.3299999999999983 0 2.306666666666665 1.0166666666666666t0.9766666666666737 2.3433333333333337v6.640000000000001q-1.3283333333333331 0-2.3049999999999997 1.0166666666666657t-0.9766666666666666 2.3416666666666686z' })
                )
            );
        }
    }]);

    return MdLocalPlay;
}(React.Component);

exports.default = MdLocalPlay;
module.exports = exports['default'];