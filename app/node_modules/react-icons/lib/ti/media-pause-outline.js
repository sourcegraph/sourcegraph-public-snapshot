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

var TiMediaPauseOutline = function (_React$Component) {
    _inherits(TiMediaPauseOutline, _React$Component);

    function TiMediaPauseOutline() {
        _classCallCheck(this, TiMediaPauseOutline);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiMediaPauseOutline).apply(this, arguments));
    }

    _createClass(TiMediaPauseOutline, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm13.333333333333334 33.333333333333336c-2.756666666666666 0-5-2.2433333333333323-5-5v-15.000000000000002c0-2.756666666666666 2.243333333333334-5 5-5s5.000000000000002 2.243333333333334 5.000000000000002 5v15.000000000000002c0 2.7566666666666677-2.2433333333333323 5-5 5z m0-21.666666666666668c-0.9199999999999999 0-1.666666666666666 0.75-1.666666666666666 1.666666666666666v15.000000000000002c0 0.9166666666666679 0.7466666666666661 1.6666666666666679 1.666666666666666 1.6666666666666679s1.666666666666666-0.75 1.666666666666666-1.6666666666666679v-15.000000000000002c0-0.9166666666666661-0.7466666666666661-1.666666666666666-1.666666666666666-1.666666666666666z m11.666666666666666 21.666666666666668c-2.7566666666666677 0-5-2.2433333333333323-5-5v-15.000000000000002c0-2.756666666666666 2.2433333333333323-5 5-5s5 2.243333333333334 5 5v15.000000000000002c0 2.7566666666666677-2.2433333333333323 5-5 5z m0-21.666666666666668c-0.9200000000000017 0-1.6666666666666679 0.75-1.6666666666666679 1.666666666666666v15.000000000000002c0 0.9166666666666679 0.7466666666666661 1.6666666666666679 1.6666666666666679 1.6666666666666679s1.6666666666666679-0.75 1.6666666666666679-1.6666666666666679v-15.000000000000002c0-0.9166666666666661-0.7466666666666661-1.666666666666666-1.6666666666666679-1.666666666666666z' })
                )
            );
        }
    }]);

    return TiMediaPauseOutline;
}(React.Component);

exports.default = TiMediaPauseOutline;
module.exports = exports['default'];