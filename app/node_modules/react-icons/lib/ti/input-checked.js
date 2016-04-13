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

var TiInputChecked = function (_React$Component) {
    _inherits(TiInputChecked, _React$Component);

    function TiInputChecked() {
        _classCallCheck(this, TiInputChecked);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiInputChecked).apply(this, arguments));
    }

    _createClass(TiInputChecked, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm26.666666666666668 31.666666666666668h-13.333333333333334c-2.756666666666666 0-5-2.2433333333333323-5-5v-13.333333333333334c0-2.756666666666666 2.243333333333334-5 5-5h8.333333333333334c0.9216666666666669 0 1.6666666666666679 0.7466666666666661 1.6666666666666679 1.666666666666666s-0.745000000000001 1.666666666666666-1.6666666666666679 1.666666666666666h-8.333333333333334c-0.9199999999999999 0-1.666666666666666 0.75-1.666666666666666 1.666666666666666v13.333333333333332c0 0.9166666666666679 0.7466666666666661 1.6666666666666679 1.666666666666666 1.6666666666666679h13.333333333333334c0.9200000000000017 0 1.6666666666666679-0.75 1.6666666666666679-1.6666666666666679v-5c0-0.9200000000000017 0.745000000000001-1.6666666666666679 1.6666666666666679-1.6666666666666679s1.6666666666666679 0.7466666666666661 1.6666666666666679 1.6666666666666679v5c0 2.7566666666666677-2.2433333333333323 5-5 5z m-4.723333333333333-6.945c-0.5833333333333321 0-1.1499999999999986-0.23333333333333428-1.5666666666666664-0.6499999999999986l-4.449999999999999-4.446666666666669c-0.8666666666666671-0.8666666666666671-0.8666666666666671-2.2749999999999986 0-3.1416666666666657s2.2766666666666673-0.8666666666666671 3.1466666666666683 0l2.3599999999999994 2.361666666666668 5.791666666666668-9.091666666666667c0.5949999999999989-1.0733333333333341 1.9499999999999993-1.461666666666666 3.0233333333333334-0.8666666666666671s1.456666666666667 1.9499999999999993 0.8616666666666681 3.0233333333333334l-7.223333333333333 11.666666666666666c-0.3383333333333347 0.6099999999999994-0.9433333333333316 1.0249999999999986-1.6333333333333329 1.1216666666666661-0.10666666666666558 0.01666666666666572-0.206666666666667 0.021666666666668277-0.3116666666666674 0.021666666666668277z' })
                )
            );
        }
    }]);

    return TiInputChecked;
}(React.Component);

exports.default = TiInputChecked;
module.exports = exports['default'];