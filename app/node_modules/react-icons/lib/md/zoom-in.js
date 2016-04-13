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

var MdZoomIn = function (_React$Component) {
    _inherits(MdZoomIn, _React$Component);

    function MdZoomIn() {
        _classCallCheck(this, MdZoomIn);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdZoomIn).apply(this, arguments));
    }

    _createClass(MdZoomIn, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm20 16.64h-3.3599999999999994v3.3599999999999994h-1.6400000000000006v-3.3599999999999994h-3.3599999999999994v-1.6400000000000006h3.3599999999999994v-3.3599999999999994h1.6400000000000006v3.3599999999999994h3.3599999999999994v1.6400000000000006z m-4.140000000000001 6.719999999999999q3.125 0 5.313333333333333-2.1883333333333326t2.1883333333333326-5.313333333333333-2.1883333333333326-5.313333333333333-5.313333333333333-2.1883333333333326-5.313333333333333 2.1883333333333326-2.1883333333333326 5.313333333333333 2.1883333333333326 5.313333333333333 5.313333333333333 2.1883333333333326z m10 0l8.283333333333331 8.283333333333331-2.5 2.5-8.283333333333331-8.283333333333331v-1.3283333333333331l-0.466666666666665-0.46999999999999886q-2.969999999999999 2.578333333333333-7.033333333333333 2.578333333333333-4.533333333333333 0-7.695-3.125t-3.165000000000001-7.654999999999999 3.166666666666668-7.693333333333333 7.691666666666666-3.166666666666668 7.656666666666666 3.166666666666666 3.125 7.693333333333333q0 4.063333333333333-2.578333333333333 7.033333333333331l0.466666666666665 0.466666666666665h1.3300000000000018z' })
                )
            );
        }
    }]);

    return MdZoomIn;
}(React.Component);

exports.default = MdZoomIn;
module.exports = exports['default'];