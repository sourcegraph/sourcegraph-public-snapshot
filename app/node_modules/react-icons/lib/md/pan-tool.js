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

var MdPanTool = function (_React$Component) {
    _inherits(MdPanTool, _React$Component);

    function MdPanTool() {
        _classCallCheck(this, MdPanTool);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdPanTool).apply(this, arguments));
    }

    _createClass(MdPanTool, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm38.36 9.14v24.21666666666667q0 2.7366666666666646-1.9916666666666671 4.689999999999998t-4.724999999999998 1.9533333333333331h-12.11q-2.8166666666666664 0-4.766666666666666-1.9533333333333331l-13.125000000000004-13.36q2.110000000000001-2.0333333333333314 2.1883333333333344-2.0333333333333314 0.5466666666666669-0.466666666666665 1.3283333333333331-0.466666666666665 0.5466666666666669 0 1.0166666666666666 0.23333333333333428l7.186666666666668 4.066666666666666v-19.84666666666667q0-1.0166666666666666 0.7416666666666671-1.7583333333333329t1.7583333333333329-0.7416666666666671 1.7583333333333329 0.7416666666666671 0.7416666666666671 1.7583333333333329v11.716666666666669h1.6383333333333319v-15.85666666666667q0-1.0933333333333335 0.7049999999999983-1.7966666666666669t1.7950000000000017-0.7033333333333331 1.8000000000000007 0.7033333333333334 0.6999999999999993 1.7966666666666666v15.86h1.6416666666666657v-14.216666666666667q0-1.0166666666666666 0.7416666666666671-1.7600000000000002t1.7583333333333329-0.7416666666666667 1.7583333333333329 0.7416666666666667 0.7416666666666671 1.7583333333333333v14.216666666666669h1.7166666666666686v-9.216666666666667q0-1.0166666666666675 0.7433333333333323-1.7583333333333329t1.759999999999998-0.7416666666666671 1.7583333333333329 0.7416666666666671 0.7433333333333323 1.7583333333333329z' })
                )
            );
        }
    }]);

    return MdPanTool;
}(React.Component);

exports.default = MdPanTool;
module.exports = exports['default'];