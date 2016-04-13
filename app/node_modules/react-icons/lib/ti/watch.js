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

var TiWatch = function (_React$Component) {
    _inherits(TiWatch, _React$Component);

    function TiWatch() {
        _classCallCheck(this, TiWatch);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiWatch).apply(this, arguments));
    }

    _createClass(TiWatch, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm20 21.666666666666668h3.333333333333332c0.9166666666666679 0 1.6666666666666679-0.75 1.6666666666666679-1.6666666666666679s-0.75-1.6666666666666679-1.6666666666666679-1.6666666666666679h-1.6666666666666679v-1.6666666666666679c0-0.9166666666666661-0.75-1.666666666666666-1.6666666666666679-1.666666666666666s-1.6666666666666679 0.75-1.6666666666666679 1.666666666666666v3.333333333333332c0 0.9166666666666679 0.75 1.6666666666666679 1.6666666666666679 1.6666666666666679z m8.333333333333336-9.825v-3.5083333333333346c0-2.756666666666667-2.2433333333333323-5-5-5h-6.666666666666668c-2.7566666666666677-4.440892098500626e-16-5 2.243333333333333-5 5v3.508333333333333c-2.0600000000000005 2.1050000000000004-3.333333333333334 4.983333333333336-3.333333333333334 8.158333333333333s1.2733333333333334 6.053333333333335 3.333333333333334 8.158333333333331v3.5083333333333364c0 2.756666666666664 2.243333333333334 5.0000000000000036 5 5.0000000000000036h6.666666666666668c2.7566666666666677 0 5-2.2433333333333323 5-5v-3.508333333333333c2.0599999999999987-2.1033333333333317 3.333333333333332-4.98 3.333333333333332-8.158333333333331s-1.2733333333333334-6.053333333333335-3.333333333333332-8.158333333333333z m-13.333333333333336-3.5083333333333346c0-0.916666666666667 0.75-1.666666666666667 1.6666666666666679-1.666666666666667h6.666666666666668c0.9166666666666679 0 1.6666666666666679 0.75 1.6666666666666679 1.666666666666667v3.0166666666666675c-1.4716666666666676-0.8550000000000004-3.176666666666666-1.3499999999999996-5-1.3499999999999996s-3.5283333333333324 0.4949999999999992-5 1.3499999999999996v-3.0166666666666675z m10 23.333333333333336c0 0.9166666666666643-0.75 1.6666666666666643-1.6666666666666679 1.6666666666666643h-6.666666666666668c-0.9166666666666661 0-1.666666666666666-0.75-1.666666666666666-1.6666666666666679v-3.0166666666666657c1.4716666666666658 0.8533333333333317 3.1766666666666676 1.3500000000000014 5.000000000000002 1.3500000000000014s3.5283333333333324-0.49666666666666615 5-1.3500000000000014v3.0166666666666657z m-5-3.333333333333332c-4.595000000000001 0-8.333333333333334-3.7383333333333333-8.333333333333334-8.333333333333332s3.7383333333333333-8.333333333333334 8.333333333333334-8.333333333333334 8.333333333333336 3.7383333333333333 8.333333333333336 8.333333333333334-3.7383333333333333 8.333333333333336-8.333333333333336 8.333333333333336z' })
                )
            );
        }
    }]);

    return TiWatch;
}(React.Component);

exports.default = TiWatch;
module.exports = exports['default'];