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

var TiBriefcase = function (_React$Component) {
    _inherits(TiBriefcase, _React$Component);

    function TiBriefcase() {
        _classCallCheck(this, TiBriefcase);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiBriefcase).apply(this, arguments));
    }

    _createClass(TiBriefcase, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm30 11.666666666666668c0-2.7566666666666677-2.2433333333333323-5-5-5h-10c-2.756666666666666-8.881784197001252e-16-5 2.2433333333333323-5 5-2.756666666666667 0-5 2.243333333333334-5 5v11.666666666666668c0 2.7566666666666677 2.243333333333334 5 5 5h20c2.7566666666666677 0 5-2.2433333333333323 5-5v-11.666666666666668c0-2.7566666666666677-2.2433333333333323-5-5-5z m-15-1.6666666666666679h10c0.9166666666666679 0 1.6666666666666679 0.75 1.6666666666666679 1.666666666666666h-13.333333333333334c0-0.9166666666666661 0.75-1.666666666666666 1.666666666666666-1.666666666666666z m16.666666666666668 18.333333333333336c0 0.9166666666666679-0.75 1.6666666666666679-1.6666666666666679 1.6666666666666679h-20c-0.9166666666666661 0-1.666666666666666-0.75-1.666666666666666-1.6666666666666679v-1.6666666666666679h23.333333333333336v1.6666666666666679z m-23.333333333333336-3.3333333333333357v-8.333333333333336c1.7763568394002505e-15-0.9166666666666661 0.7500000000000018-1.666666666666666 1.6666666666666679-1.666666666666666h20c0.9166666666666679 0 1.6666666666666679 0.75 1.6666666666666679 1.666666666666666v8.333333333333336h-23.333333333333336z m13.333333333333336-5h-3.333333333333332c-0.9166666666666679 0-1.6666666666666679 0.75-1.6666666666666679 1.6666666666666679s0.75 1.6666666666666679 1.6666666666666679 1.6666666666666679h3.333333333333332c0.9166666666666679 0 1.6666666666666679-0.75 1.6666666666666679-1.6666666666666679s-0.75-1.6666666666666679-1.6666666666666679-1.6666666666666679z' })
                )
            );
        }
    }]);

    return TiBriefcase;
}(React.Component);

exports.default = TiBriefcase;
module.exports = exports['default'];