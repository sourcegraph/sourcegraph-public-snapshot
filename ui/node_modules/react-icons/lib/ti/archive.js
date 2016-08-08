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

var TiArchive = function (_React$Component) {
    _inherits(TiArchive, _React$Component);

    function TiArchive() {
        _classCallCheck(this, TiArchive);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiArchive).apply(this, arguments));
    }

    _createClass(TiArchive, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm21.666666666666668 20h-5c-0.46000000000000085 0-0.8333333333333339 0.37333333333333485-0.8333333333333339 0.8333333333333321s0.3733333333333331 0.8333333333333321 0.8333333333333339 0.8333333333333321h5c0.46000000000000085 0 0.8333333333333321-0.37333333333333485 0.8333333333333321-0.8333333333333321s-0.37333333333333485-0.8333333333333321-0.8333333333333321-0.8333333333333321z m11.666666666666668-11.666666666666666h-28.333333333333336c-0.9216666666666669 0-1.666666666666667 0.7466666666666661-1.666666666666667 1.666666666666666s0.7450000000000001 1.666666666666666 1.666666666666667 1.666666666666666h28.333333333333336c0.9216666666666669 0 1.6666666666666643-0.7466666666666661 1.6666666666666643-1.666666666666666s-0.7449999999999974-1.666666666666666-1.6666666666666643-1.666666666666666z m-3.3333333333333357 5h-21.666666666666664c-0.9216666666666686 0-1.6666666666666687 0.7466666666666661-1.6666666666666687 1.666666666666666v13.333333333333336c0 2.7566666666666677 2.243333333333333 5 5.000000000000001 5h15c2.7566666666666677 0 5-2.2433333333333323 5-5v-13.333333333333336c0-0.9199999999999999-0.745000000000001-1.666666666666666-1.6666666666666679-1.666666666666666z m-3.333333333333332 16.666666666666664h-15c-0.9199999999999999 0-1.666666666666666-0.75-1.666666666666666-1.6666666666666679v-11.666666666666668h18.333333333333336v11.666666666666668c0 0.9166666666666679-0.7466666666666661 1.6666666666666679-1.6666666666666679 1.6666666666666679z' })
                )
            );
        }
    }]);

    return TiArchive;
}(React.Component);

exports.default = TiArchive;
module.exports = exports['default'];