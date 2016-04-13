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

var MdAvTimer = function (_React$Component) {
    _inherits(MdAvTimer, _React$Component);

    function MdAvTimer() {
        _classCallCheck(this, MdAvTimer);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdAvTimer).apply(this, arguments));
    }

    _createClass(MdAvTimer, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm10 20q0-0.7033333333333331 0.4666666666666668-1.1716666666666669t1.1733333333333338-0.466666666666665 1.211666666666666 0.466666666666665 0.5099999999999998 1.1716666666666669-0.5083333333333329 1.1716666666666669-1.209999999999999 0.466666666666665-1.1716666666666669-0.466666666666665-0.47166666666666757-1.1716666666666669z m20 0q0 0.7033333333333331-0.466666666666665 1.1716666666666669t-1.173333333333332 0.466666666666665-1.211666666666666-0.466666666666665-0.5100000000000016-1.1716666666666669 0.5083333333333329-1.1716666666666669 1.2100000000000009-0.466666666666665 1.1716666666666669 0.466666666666665 0.471666666666664 1.1716666666666669z m-11.64-15h1.6400000000000006q6.25 0 10.625 4.375t4.375 10.625-4.375 10.625-10.625 4.375-10.625-4.375-4.375-10.625q0-7.5 6.016666666666666-11.953333333333333v-0.07999999999999918l11.326666666666668 11.329999999999998-2.3433333333333337 2.3433333333333337-9.063333333333333-8.983333333333333q-2.578333333333333 3.200000000000001-2.578333333333333 7.341666666666667 0 4.841666666666669 3.4000000000000004 8.240000000000002t8.24 3.3999999999999986 8.240000000000002-3.3999999999999986 3.4016666666666673-8.238333333333337q0-4.376666666666667-2.8916666666666657-7.658333333333333t-7.111666666666672-3.908333333333333v3.205h-3.283333333333335v-6.638333333333334z m0 23.36q0-0.7033333333333331 0.466666666666665-1.211666666666666t1.173333333333332-0.5100000000000016 1.1716666666666669 0.5083333333333329 0.466666666666665 1.2100000000000009-0.466666666666665 1.1716666666666669-1.1716666666666633 0.47166666666666757-1.1716666666666669-0.466666666666665-0.466666666666665-1.173333333333332z' })
                )
            );
        }
    }]);

    return MdAvTimer;
}(React.Component);

exports.default = MdAvTimer;
module.exports = exports['default'];