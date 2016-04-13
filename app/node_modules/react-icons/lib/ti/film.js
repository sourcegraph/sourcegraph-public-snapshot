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

var TiFilm = function (_React$Component) {
    _inherits(TiFilm, _React$Component);

    function TiFilm() {
        _classCallCheck(this, TiFilm);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiFilm).apply(this, arguments));
    }

    _createClass(TiFilm, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm13.333333333333334 13.333333333333334v11.666666666666666h13.333333333333334v-11.666666666666666h-13.333333333333334z m11.666666666666666 10.000000000000002h-10v-8.333333333333336h10v8.333333333333336z m3.3333333333333357-20.000000000000004h-5v3.333333333333335h-6.666666666666668v-3.3333333333333335h-5c-2.7566666666666677 0-5.000000000000001 2.2433333333333336-5.000000000000001 5v21.666666666666664c0 2.7566666666666677 2.243333333333333 5 5.000000000000001 5h5v-3.333333333333332h6.666666666666668v3.333333333333332h5c2.7566666666666677 0 5-2.2433333333333323 5-5v-21.666666666666664c0-2.7566666666666686-2.2433333333333323-5.000000000000002-5-5.000000000000002z m1.6666666666666679 6.666666666666668c-0.9216666666666669 0-1.6666666666666679 0.7449999999999992-1.6666666666666679 1.666666666666666s0.745000000000001 1.666666666666666 1.6666666666666679 1.666666666666666v1.666666666666666c-0.9216666666666669 0-1.6666666666666679 0.7449999999999992-1.6666666666666679 1.666666666666666s0.745000000000001 1.6666666666666679 1.6666666666666679 1.6666666666666679v1.6666666666666679c-0.9216666666666669 0-1.6666666666666679 0.745000000000001-1.6666666666666679 1.6666666666666679s0.745000000000001 1.6666666666666679 1.6666666666666679 1.6666666666666679v1.6666666666666679c-0.9216666666666669 0-1.6666666666666679 0.745000000000001-1.6666666666666679 1.6666666666666679s0.745000000000001 1.6666666666666679 1.6666666666666679 1.6666666666666679v1.6666666666666679c0 0.9166666666666679-0.7466666666666661 1.6666666666666679-1.6666666666666679 1.6666666666666679h-1.6666666666666679v-3.333333333333332h-13.333333333333334v3.333333333333332h-1.666666666666666c-0.9199999999999999 0-1.666666666666666-0.75-1.666666666666666-1.6666666666666679v-1.6666666666666679c0.9216666666666669 0 1.666666666666666-0.745000000000001 1.666666666666666-1.6666666666666679s-0.7449999999999992-1.6666666666666679-1.666666666666666-1.6666666666666679v-1.6666666666666679c0.9216666666666669 0 1.666666666666666-0.745000000000001 1.666666666666666-1.6666666666666679s-0.7449999999999992-1.6666666666666679-1.666666666666666-1.6666666666666679v-1.6666666666666679c0.9216666666666669 0 1.666666666666666-0.745000000000001 1.666666666666666-1.6666666666666679s-0.7449999999999992-1.666666666666666-1.666666666666666-1.666666666666666v-1.6666666666666643c0.9216666666666669 0 1.666666666666666-0.7449999999999992 1.666666666666666-1.666666666666666s-0.7449999999999992-1.666666666666666-1.666666666666666-1.666666666666666v-1.6666666666666679c0-0.916666666666667 0.7466666666666661-1.666666666666667 1.666666666666666-1.666666666666667h1.666666666666666v3.333333333333333h13.333333333333334v-3.333333333333333h1.6666666666666679c0.9200000000000017 0 1.6666666666666679 0.75 1.6666666666666679 1.666666666666667v1.666666666666666z' })
                )
            );
        }
    }]);

    return TiFilm;
}(React.Component);

exports.default = TiFilm;
module.exports = exports['default'];