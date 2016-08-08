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

var MdRingVolume = function (_React$Component) {
    _inherits(MdRingVolume, _React$Component);

    function MdRingVolume() {
        _classCallCheck(this, MdRingVolume);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdRingVolume).apply(this, arguments));
    }

    _createClass(MdRingVolume, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm10.703333333333333 16.328333333333337q-5.783333333333333-5.783333333333337-5.9366666666666665-5.86166666666667l2.3400000000000007-2.416666666666666 5.9399999999999995 5.933333333333334z m10.936666666666667-12.966666666666667v8.280000000000001h-3.2833333333333314v-8.283333333333333h3.2833333333333314z m13.593333333333334 7.104999999999997l-5.933333333333337 5.863333333333335-2.3466666666666605-2.3466666666666676 5.938333333333336-5.936666666666667z m4.299999999999997 17.346666666666664q0.46666666666666856 0.466666666666665 0.46666666666666856 1.1716666666666669t-0.46666666666666856 1.173333333333332l-4.141666666666666 4.140000000000001q-0.46666666666666856 0.46666666666666856-1.1716666666666669 0.46666666666666856t-1.1716666666666669-0.46666666666666856q-2.1900000000000013-2.0333333333333314-4.454999999999998-3.125-0.9383333333333326-0.39000000000000057-0.9383333333333326-1.4833333333333343v-5.158333333333335q-3.5933333333333337-1.1716666666666669-7.656666666666666-1.1716666666666669t-7.656666666666666 1.1716666666666669v5.156666666666666q0 1.1716666666666669-0.9383333333333326 1.5633333333333326-2.5 1.1716666666666633-4.453333333333334 3.0466666666666704-0.4666666666666668 0.46666666666666856-1.17 0.46666666666666856t-1.1716666666666669-0.46666666666666856l-4.141666666666668-4.141666666666662q-0.4666666666666668-0.46666666666666856-0.4666666666666668-1.1733333333333356t0.46666666666666673-1.1700000000000017q8.206666666666667-7.813333333333333 19.533333333333335-7.813333333333333t19.53333333333333 7.813333333333333z' })
                )
            );
        }
    }]);

    return MdRingVolume;
}(React.Component);

exports.default = MdRingVolume;
module.exports = exports['default'];