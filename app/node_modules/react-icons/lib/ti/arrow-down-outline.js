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

var TiArrowDownOutline = function (_React$Component) {
    _inherits(TiArrowDownOutline, _React$Component);

    function TiArrowDownOutline() {
        _classCallCheck(this, TiArrowDownOutline);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiArrowDownOutline).apply(this, arguments));
    }

    _createClass(TiArrowDownOutline, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm20 35.52l-11.866666666666667-11.866666666666667c-1.9500000000000002-1.9499999999999993-1.9500000000000002-5.123333333333335 0-7.071666666666665 1.820000000000002-1.8250000000000046 4.960000000000001-1.8983333333333352 6.866666666666667-0.19333333333333513v-8.055000000000001c0-2.756666666666667 2.2433333333333323-5 5-5s5 2.243333333333334 5 5v8.056666666666667c1.9050000000000011-1.705 5.045000000000002-1.6333333333333329 6.866666666666667 0.19166666666666643 1.9500000000000028 1.9466666666666654 1.9500000000000028 5.116666666666667 0 7.066666666666666l-11.866666666666667 11.873333333333335z m-8.333333333333332-17.07c-0.4466666666666672 0-0.8633333333333333 0.173333333333332-1.1783333333333328 0.4883333333333333-0.6500000000000004 0.6499999999999986-0.6500000000000004 1.7049999999999983 0 2.3566666666666656l9.511666666666665 9.51166666666667 9.511666666666667-9.511666666666667c0.6499999999999986-0.6499999999999986 0.6499999999999986-1.706666666666667 0-2.3566666666666656-0.6333333333333329-0.6333333333333329-1.7250000000000014-0.6333333333333329-2.3566666666666656 0l-5.488333333333333 5.48833333333333v-16.093333333333334c0-0.9166666666666687-0.7466666666666661-1.6666666666666687-1.6666666666666679-1.6666666666666687s-1.6666666666666679 0.75-1.6666666666666679 1.666666666666667v16.093333333333334l-5.488333333333333-5.488333333333333c-0.31666666666666643-0.31666666666666643-0.7333333333333343-0.4883333333333333-1.1783333333333328-0.4883333333333333z' })
                )
            );
        }
    }]);

    return TiArrowDownOutline;
}(React.Component);

exports.default = TiArrowDownOutline;
module.exports = exports['default'];