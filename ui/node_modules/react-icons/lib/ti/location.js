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

var TiLocation = function (_React$Component) {
    _inherits(TiLocation, _React$Component);

    function TiLocation() {
        _classCallCheck(this, TiLocation);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiLocation).apply(this, arguments));
    }

    _createClass(TiLocation, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm29.428333333333335 8.840000000000002c-5.206666666666667-5.121666666666667-13.65-5.121666666666667-18.855 0s-5.206666666666667 13.428333333333336 0 18.549999999999997l9.426666666666666 9.27666666666667 9.428333333333335-9.276666666666667c5.2066666666666706-5.121666666666666 5.2066666666666706-13.426666666666666 0-18.55z m-9.428333333333335 13.659999999999998c-1.1133333333333333 0-2.1583333333333314-0.43333333333333357-2.9466666666666654-1.2216666666666676-1.625-1.625-1.625-4.266666666666666 0-5.893333333333334 0.7866666666666653-0.7833333333333332 1.8333333333333321-1.2166666666666668 2.9466666666666654-1.2166666666666668s2.16 0.43333333333333357 2.9466666666666654 1.2166666666666668c1.625 1.6266666666666652 1.625 4.271666666666668 0 5.895-0.7866666666666653 0.7866666666666653-1.8333333333333321 1.2199999999999989-2.9466666666666654 1.2199999999999989z' })
                )
            );
        }
    }]);

    return TiLocation;
}(React.Component);

exports.default = TiLocation;
module.exports = exports['default'];