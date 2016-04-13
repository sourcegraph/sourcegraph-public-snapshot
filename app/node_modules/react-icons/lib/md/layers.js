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

var MdLayers = function (_React$Component) {
    _inherits(MdLayers, _React$Component);

    function MdLayers() {
        _classCallCheck(this, MdLayers);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdLayers).apply(this, arguments));
    }

    _createClass(MdLayers, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm20 26.64q-0.7033333333333331-0.5466666666666669-6.366666666666667-4.961666666666666t-8.633333333333333-6.678333333333335l15-11.641666666666667 15 11.641666666666667q-2.9666666666666686 2.2633333333333354-8.593333333333334 6.638333333333335t-6.406666666666666 5z m0 4.300000000000001l12.266666666666666-9.611666666666668 2.7333333333333343 2.110000000000003-15 11.639999999999997-15-11.64 2.7333333333333334-2.1099999999999994z' })
                )
            );
        }
    }]);

    return MdLayers;
}(React.Component);

exports.default = MdLayers;
module.exports = exports['default'];