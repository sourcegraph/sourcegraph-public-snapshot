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

var MdBeenhere = function (_React$Component) {
    _inherits(MdBeenhere, _React$Component);

    function MdBeenhere() {
        _classCallCheck(this, MdBeenhere);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdBeenhere).apply(this, arguments));
    }

    _createClass(MdBeenhere, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm16.64 26.64l15-15-2.3433333333333337-2.3433333333333337-12.656666666666666 12.656666666666666-5.9399999999999995-5.936666666666667-2.341666666666667 2.3416666666666686z m15-25q1.3283333333333331 0 2.3433333333333337 1.0166666666666666t1.0166666666666657 2.34v21.561666666666667q0 1.6400000000000006-1.4866666666666646 2.7333333333333343l-13.516666666666666 9.066666666666666-13.513333333333335-9.06666666666667q-1.4833333333333343-1.086666666666666-1.4833333333333343-2.728333333333328v-21.563333333333336q0-1.33 1.0133333333333336-2.345t2.3433333333333337-1.0166666666666666h23.28333333333333z' })
                )
            );
        }
    }]);

    return MdBeenhere;
}(React.Component);

exports.default = MdBeenhere;
module.exports = exports['default'];