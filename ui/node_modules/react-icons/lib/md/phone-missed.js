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

var MdPhoneMissed = function (_React$Component) {
    _inherits(MdPhoneMissed, _React$Component);

    function MdPhoneMissed() {
        _classCallCheck(this, MdPhoneMissed);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdPhoneMissed).apply(this, arguments));
    }

    _createClass(MdPhoneMissed, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm39.53333333333333 27.813333333333333q0.46666666666666856 0.466666666666665 0.46666666666666856 1.1716666666666669t-0.46666666666666856 1.173333333333332l-4.141666666666666 4.140000000000001q-0.46666666666666856 0.46666666666666856-1.1716666666666669 0.46666666666666856t-1.1716666666666669-0.46666666666666856q-2.1900000000000013-2.0333333333333314-4.454999999999998-3.125-0.9383333333333326-0.39000000000000057-0.9383333333333326-1.4833333333333343v-5.158333333333335q-3.5933333333333337-1.1716666666666669-7.656666666666666-1.1716666666666669t-7.656666666666666 1.1716666666666669v5.156666666666666q0 1.1716666666666669-0.9383333333333326 1.5633333333333326-2.5 1.1716666666666633-4.453333333333334 3.0466666666666704-0.4666666666666668 0.46666666666666856-1.17 0.46666666666666856t-1.1716666666666669-0.46666666666666856l-4.141666666666668-4.141666666666662q-0.4666666666666668-0.46666666666666856-0.4666666666666668-1.1733333333333356t0.46666666666666673-1.1700000000000017q8.206666666666667-7.813333333333333 19.533333333333335-7.813333333333333t19.53333333333333 7.813333333333333z m-28.674999999999997-18.673333333333332v5.859999999999999h-2.5000000000000018v-10h10.000000000000002v2.5h-5.858333333333334l7.5 7.5 10-10 1.6400000000000006 1.6400000000000006-11.64 11.716666666666665z' })
                )
            );
        }
    }]);

    return MdPhoneMissed;
}(React.Component);

exports.default = MdPhoneMissed;
module.exports = exports['default'];