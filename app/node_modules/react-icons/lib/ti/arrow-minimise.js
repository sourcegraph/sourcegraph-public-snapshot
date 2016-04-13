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

var TiArrowMinimise = function (_React$Component) {
    _inherits(TiArrowMinimise, _React$Component);

    function TiArrowMinimise() {
        _classCallCheck(this, TiArrowMinimise);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiArrowMinimise).apply(this, arguments));
    }

    _createClass(TiArrowMinimise, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm10.200000000000001 21.666666666666668c-0.9199999999999999 0-1.666666666666666 0.7466666666666661-1.666666666666666 1.6666666666666679s0.7466666666666661 1.6666666666666679 1.666666666666666 1.6666666666666679h2.4433333333333334l-5.488333333333334 5.488333333333333c-0.6500000000000004 0.6499999999999986-0.6500000000000004 1.7049999999999983 0 2.356666666666669 0.3250000000000002 0.32500000000000284 0.75 0.48833333333333684 1.1783333333333337 0.48833333333333684s0.8533333333333335-0.163333333333334 1.1783333333333328-0.48833333333333684l5.690000000000001-5.690000000000001v2.8449999999999953c0 0.9200000000000017 0.7466666666666661 1.6666666666666679 1.6666666666666679 1.6666666666666679s1.466666666666665-0.7466666666666661 1.466666666666665-1.6666666666666679v-8.333333333333336h-8.135z m1.4666666666666668-3.333333333333332c0.9199999999999999 0 1.666666666666666-0.7466666666666661 1.666666666666666-1.6666666666666679v-3.333333333333334h3.333333333333334c0.9216666666666669 0 1.6666666666666679-0.7466666666666661 1.6666666666666679-1.666666666666666s-0.745000000000001-1.666666666666666-1.6666666666666679-1.666666666666666h-6.666666666666668v6.666666666666666c0 0.9200000000000017 0.7449999999999992 1.6666666666666679 1.666666666666666 1.6666666666666679z m16.666666666666668 3.333333333333332c-0.9216666666666669 0-1.6666666666666679 0.7466666666666661-1.6666666666666679 1.6666666666666679v3.333333333333332h-3.333333333333332c-0.9216666666666669 0-1.6666666666666679 0.7466666666666661-1.6666666666666679 1.6666666666666679s0.745000000000001 1.6666666666666679 1.6666666666666679 1.6666666666666679h6.666666666666668v-6.666666666666668c0-0.9200000000000017-0.745000000000001-1.6666666666666679-1.6666666666666679-1.6666666666666679z m2.1549999999999976-14.511666666666667l-5.488333333333333 5.488333333333333v-2.6433333333333344c0-0.9199999999999999-0.745000000000001-1.666666666666666-1.6666666666666679-1.666666666666666s-1.6666666666666679 0.7466666666666661-1.6666666666666679 1.666666666666666v8.333333333333336h8.333333333333336c0.9200000000000017 0 1.6666666666666679-0.7466666666666661 1.6666666666666679-1.6666666666666679s-0.745000000000001-1.666666666666666-1.6666666666666679-1.666666666666666h-2.6433333333333344l5.488333333333333-5.486666666666666c0.6499999999999986-0.6500000000000004 0.6499999999999986-1.705 0-2.3566666666666665s-1.7049999999999983-0.6533333333333333-2.3566666666666656 0z' })
                )
            );
        }
    }]);

    return TiArrowMinimise;
}(React.Component);

exports.default = TiArrowMinimise;
module.exports = exports['default'];