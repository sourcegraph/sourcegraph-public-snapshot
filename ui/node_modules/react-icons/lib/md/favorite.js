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

var MdFavorite = function (_React$Component) {
    _inherits(MdFavorite, _React$Component);

    function MdFavorite() {
        _classCallCheck(this, MdFavorite);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdFavorite).apply(this, arguments));
    }

    _createClass(MdFavorite, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm20 35.54666666666667l-2.421666666666667-2.1883333333333326q-4.140000000000001-3.75-6.016666666666666-5.546666666666667t-4.178333333333335-4.453333333333333-3.163333333333333-4.805-0.8600000000000003-4.413333333333334q0-3.828333333333333 2.616666666666667-6.483333333333333t6.523333333333333-2.66q4.533333333333335 0 7.5 3.5133333333333345 2.9666666666666686-3.5166666666666666 7.5-3.5166666666666666 3.9066666666666663 0 6.523333333333333 2.658333333333333t2.616666666666667 6.483333333333333q0 3.0500000000000007-2.0333333333333314 6.330000000000002t-4.411666666666665 5.704999999999998-7.773333333333333 7.266666666666666z' })
                )
            );
        }
    }]);

    return MdFavorite;
}(React.Component);

exports.default = MdFavorite;
module.exports = exports['default'];