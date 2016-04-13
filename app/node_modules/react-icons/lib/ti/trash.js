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

var TiTrash = function (_React$Component) {
    _inherits(TiTrash, _React$Component);

    function TiTrash() {
        _classCallCheck(this, TiTrash);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiTrash).apply(this, arguments));
    }

    _createClass(TiTrash, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm30 11.666666666666668h-1.6666666666666679v-1.6666666666666679c0-1.8399999999999999-1.4933333333333323-3.333333333333334-3.333333333333332-3.333333333333334h-11.666666666666666c-1.8399999999999999 0-3.333333333333334 1.493333333333334-3.333333333333334 3.333333333333334v1.666666666666666h-1.666666666666666c-0.9199999999999999 0-1.666666666666667 0.7466666666666661-1.666666666666667 1.666666666666666s0.746666666666667 1.666666666666666 1.666666666666667 1.666666666666666v13.333333333333334c0 3.676666666666666 2.99 6.666666666666668 6.666666666666666 6.666666666666668h8.333333333333336c3.676666666666666 0 6.666666666666668-2.990000000000002 6.666666666666668-6.666666666666668v-13.333333333333332c0.9200000000000017 0 1.6666666666666679-0.7466666666666661 1.6666666666666679-1.666666666666666s-0.7466666666666661-1.666666666666666-1.6666666666666679-1.666666666666666z m-16.666666666666664-1.6666666666666679h11.666666666666664v1.666666666666666h-11.666666666666666v-1.666666666666666z m13.333333333333332 18.333333333333336c0 1.8399999999999999-1.4933333333333323 3.333333333333332-3.333333333333332 3.333333333333332h-8.333333333333336c-1.8399999999999999 0-3.333333333333334-1.4933333333333323-3.333333333333334-3.333333333333332v-13.333333333333336h14.999999999999998v13.333333333333336z m-12.5-10.833333333333336c-0.4583333333333339 0-0.8333333333333339 0.375-0.8333333333333339 0.8333333333333321v10c0 0.45833333333333215 0.375 0.8333333333333321 0.8333333333333339 0.8333333333333321s0.8333333333333339-0.375 0.8333333333333339-0.8333333333333321v-10c0-0.45833333333333215-0.375-0.8333333333333321-0.8333333333333339-0.8333333333333321z m3.333333333333332 0c-0.45833333333333215 0-0.8333333333333321 0.375-0.8333333333333321 0.8333333333333321v10c0 0.45833333333333215 0.375 0.8333333333333321 0.8333333333333321 0.8333333333333321s0.8333333333333321-0.375 0.8333333333333321-0.8333333333333321v-10c0-0.45833333333333215-0.375-0.8333333333333321-0.8333333333333321-0.8333333333333321z m3.3333333333333357 0c-0.45833333333333215 0-0.8333333333333321 0.375-0.8333333333333321 0.8333333333333321v10c0 0.45833333333333215 0.375 0.8333333333333321 0.8333333333333321 0.8333333333333321s0.8333333333333321-0.375 0.8333333333333321-0.8333333333333321v-10c0-0.45833333333333215-0.375-0.8333333333333321-0.8333333333333321-0.8333333333333321z m3.333333333333332 0c-0.45833333333333215 0-0.8333333333333321 0.375-0.8333333333333321 0.8333333333333321v10c0 0.45833333333333215 0.375 0.8333333333333321 0.8333333333333321 0.8333333333333321s0.8333333333333321-0.375 0.8333333333333321-0.8333333333333321v-10c0-0.45833333333333215-0.375-0.8333333333333321-0.8333333333333321-0.8333333333333321z' })
                )
            );
        }
    }]);

    return TiTrash;
}(React.Component);

exports.default = TiTrash;
module.exports = exports['default'];