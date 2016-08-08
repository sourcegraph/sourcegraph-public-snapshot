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

var TiNotesOutline = function (_React$Component) {
    _inherits(TiNotesOutline, _React$Component);

    function TiNotesOutline() {
        _classCallCheck(this, TiNotesOutline);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiNotesOutline).apply(this, arguments));
    }

    _createClass(TiNotesOutline, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm30.540000000000003 7.278333333333333c-0.6133333333333333-0.54-1.3999999999999986-0.833333333333333-2.206666666666667-0.833333333333333-0.13666666666666671 0-0.2749999999999986 0.008333333333333748-0.413333333333334 0.026666666666666394l-15 2.083333333333333c-1.6666666666666679 0.20833333333333393-2.9200000000000017 1.625-2.9200000000000017 3.3066666666666666v10.183333333333335c-2.91 0.9216666666666633-5 3.393333333333331-5 6.288333333333334 0 3.674999999999997 3.366666666666667 6.666666666666664 7.5 6.666666666666664 2.8249999999999993 0 5.283333333333331-1.403333333333336 6.561666666666667-3.4633333333333347 1.3383333333333347 1.1133333333333333 3.133333333333333 1.7966666666666704 5.105 1.7966666666666704 4.133333333333333 0 7.5-2.991666666666667 7.5-6.666666666666668v-16.888333333333335c0-0.9566666666666652-0.41000000000000014-1.8666666666666654-1.1266666666666652-2.4999999999999982z m-12.206666666666667 19.388333333333335v-7.093333333333334l5-0.75v2.8949999999999996c-2.8216666666666654 0.3500000000000014-5 2.4333333333333336-5 4.949999999999999z m10 0c0 1.8399999999999999-1.8666666666666671 3.333333333333332-4.166666666666668 3.333333333333332s-4.166666666666668-1.4933333333333323-4.166666666666668-3.333333333333332 1.8666666666666671-3.333333333333332 4.166666666666668-3.333333333333332c0.2866666666666653 0 0.5633333333333326 0.023333333333333428 0.8333333333333321 0.06666666666666643v-6.511666666666667l-8.333333333333336 1.25v10.195c0 1.8399999999999999-1.8666666666666671 3.333333333333332-4.166666666666668 3.333333333333332s-4.1666666666666625-1.4933333333333358-4.1666666666666625-3.333333333333332 1.8666666666666671-3.333333333333332 4.166666666666666-3.333333333333332c0.28666666666666707 0 0.5633333333333326 0.023333333333333428 0.8333333333333339 0.06666666666666643v-13.203333333333335l15.000000000000002-2.083333333333334v16.886666666666667z' })
                )
            );
        }
    }]);

    return TiNotesOutline;
}(React.Component);

exports.default = TiNotesOutline;
module.exports = exports['default'];