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

var MdVoicemail = function (_React$Component) {
    _inherits(MdVoicemail, _React$Component);

    function MdVoicemail() {
        _classCallCheck(this, MdVoicemail);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdVoicemail).apply(this, arguments));
    }

    _createClass(MdVoicemail, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm30.86 25q2.421666666666667 0 4.100000000000001-1.7166666666666686t1.68333333333333-4.141666666666666-1.68333333333333-4.100000000000001-4.100000000000001-1.6833333333333336-4.140000000000001 1.6833333333333336-1.7166666666666686 4.100000000000001 1.7166666666666686 4.140000000000001 4.140000000000001 1.7183333333333337z m-21.72 0q2.421666666666667 0 4.140000000000001-1.7166666666666686t1.7166666666666668-4.141666666666666-1.7166666666666668-4.100000000000001-4.140000000000001-1.6833333333333336-4.1000000000000005 1.6833333333333336-1.6833333333333336 4.100000000000001 1.6833333333333336 4.140000000000001 4.098333333333334 1.7183333333333337z m21.720000000000002-15q3.8283333333333367 0 6.483333333333331 2.6566666666666663t2.6583333333333314 6.483333333333334-2.655000000000001 6.524999999999999-6.483333333333334 2.693333333333335h-21.72333333333333q-3.828333333333333 0-6.483333333333333-2.6950000000000003t-2.656666666666667-6.523333333333333 2.6566666666666667-6.483333333333334 6.483333333333334-2.6566666666666663 6.526666666666667 2.656666666666668 2.693333333333335 6.4833333333333325q0 3.361666666666668-2.1099999999999994 5.861666666666665h7.5q-2.1099999999999994-2.5-2.1099999999999994-5.858333333333334 0-3.828333333333333 2.6950000000000003-6.483333333333334t6.524999999999995-2.6599999999999966z' })
                )
            );
        }
    }]);

    return MdVoicemail;
}(React.Component);

exports.default = MdVoicemail;
module.exports = exports['default'];