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

var MdNoSim = function (_React$Component) {
    _inherits(MdNoSim, _React$Component);

    function MdNoSim() {
        _classCallCheck(this, MdNoSim);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdNoSim).apply(this, arguments));
    }

    _createClass(MdNoSim, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm6.093333333333334 6.483333333333333l29.14 29.066666666666663-2.1883333333333326 2.1866666666666674-3.125-3.203333333333333q-0.9383333333333326 0.46666666666666856-1.5633333333333326 0.46666666666666856h-16.71666666666667q-1.33 0-2.3066666666666666-1.0133333333333354t-0.9766666666666666-2.3433333333333337v-18.671666666666663l-4.373333333333333-4.375z m25.546666666666667 1.876666666666666v19.453333333333337l-18.906666666666666-18.906666666666666 3.9066666666666663-3.90666666666667h11.716666666666669q1.3299999999999983 0 2.306666666666665 1.0166666666666666t0.9766666666666666 2.3416666666666677z' })
                )
            );
        }
    }]);

    return MdNoSim;
}(React.Component);

exports.default = MdNoSim;
module.exports = exports['default'];