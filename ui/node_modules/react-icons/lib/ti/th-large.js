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

var TiThLarge = function (_React$Component) {
    _inherits(TiThLarge, _React$Component);

    function TiThLarge() {
        _classCallCheck(this, TiThLarge);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiThLarge).apply(this, arguments));
    }

    _createClass(TiThLarge, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm13.333333333333334 5h-3.333333333333334c-1.375 0-2.625 0.5616666666666665-3.533333333333333 1.4666666666666668s-1.4666666666666668 2.16-1.4666666666666668 3.533333333333333v3.333333333333334c0 1.375 0.5616666666666665 2.625 1.4666666666666668 3.533333333333333s2.16 1.4666666666666686 3.533333333333333 1.4666666666666686h3.333333333333334c1.375 0 2.625-0.5616666666666674 3.533333333333333-1.4666666666666686s1.4666666666666686-2.16 1.4666666666666686-3.533333333333333v-3.333333333333334c0-1.375-0.5616666666666674-2.625-1.4666666666666686-3.533333333333333s-2.16-1.4666666666666668-3.533333333333333-1.4666666666666668z m16.666666666666664 0h-3.333333333333332c-1.375 0-2.625 0.5616666666666665-3.533333333333335 1.4666666666666668s-1.466666666666665 2.16-1.466666666666665 3.533333333333333v3.333333333333334c0 1.375 0.5616666666666674 2.625 1.466666666666665 3.533333333333333s2.16 1.4666666666666686 3.533333333333335 1.4666666666666686h3.333333333333332c1.375 0 2.625-0.5616666666666674 3.5333333333333314-1.4666666666666686s1.4666666666666686-2.16 1.4666666666666686-3.533333333333333v-3.333333333333334c0-1.375-0.5616666666666674-2.625-1.4666666666666686-3.533333333333333s-2.159999999999993-1.4666666666666668-3.5333333333333314-1.4666666666666668z m-16.666666666666664 16.666666666666668h-3.3333333333333357c-1.375 0-2.625 0.5616666666666674-3.533333333333333 1.466666666666665s-1.4666666666666668 2.16-1.4666666666666668 3.533333333333335v3.333333333333332c0 1.375 0.5616666666666665 2.625 1.4666666666666668 3.5333333333333314s2.16 1.4666666666666686 3.533333333333333 1.4666666666666686h3.333333333333334c1.375 0 2.625-0.5616666666666674 3.533333333333333-1.4666666666666686s1.4666666666666686-2.159999999999993 1.4666666666666686-3.5333333333333314v-3.333333333333332c0-1.375-0.5616666666666674-2.625-1.4666666666666686-3.533333333333335s-2.16-1.466666666666665-3.533333333333333-1.466666666666665z m16.666666666666664 0h-3.333333333333332c-1.375 0-2.625 0.5616666666666674-3.533333333333335 1.466666666666665s-1.466666666666665 2.16-1.466666666666665 3.533333333333335v3.333333333333332c0 1.375 0.5616666666666674 2.625 1.466666666666665 3.5333333333333314s2.16 1.4666666666666686 3.533333333333335 1.4666666666666686h3.333333333333332c1.375 0 2.625-0.5616666666666674 3.5333333333333314-1.4666666666666686s1.4666666666666686-2.159999999999993 1.4666666666666686-3.5333333333333314v-3.333333333333332c0-1.375-0.5616666666666674-2.625-1.4666666666666686-3.533333333333335s-2.159999999999993-1.466666666666665-3.5333333333333314-1.466666666666665z' })
                )
            );
        }
    }]);

    return TiThLarge;
}(React.Component);

exports.default = TiThLarge;
module.exports = exports['default'];