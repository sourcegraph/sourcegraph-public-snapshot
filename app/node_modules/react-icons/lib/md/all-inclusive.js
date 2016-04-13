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

var MdAllInclusive = function (_React$Component) {
    _inherits(MdAllInclusive, _React$Component);

    function MdAllInclusive() {
        _classCallCheck(this, MdAllInclusive);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdAllInclusive).apply(this, arguments));
    }

    _createClass(MdAllInclusive, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm31.016666666666666 11.016666666666667q3.75 0 6.366666666666667 2.6549999999999994t2.616666666666667 6.328333333333333-2.616666666666667 6.328333333333333-6.366666666666667 2.6566666666666663-6.408333333333335-2.655000000000001l-2.1083333333333307-1.8766666666666616 2.5-2.186666666666671 1.9499999999999993 1.6383333333333319q1.7199999999999989 1.716666666666665 4.066666666666666 1.716666666666665t3.9833333333333343-1.6383333333333283 1.6400000000000006-3.9833333333333343-1.6400000000000006-3.9883333333333333-3.9833333333333343-1.6400000000000006-3.986666666666668 1.6400000000000006q-5.546666666666667 4.843333333333334-7.033333333333335 6.25l-4.686666666666667 4.140000000000001q-2.5766666666666627 2.581666666666667-6.326666666666663 2.581666666666667t-6.366666666666666-2.6583333333333314-2.6166666666666663-6.328333333333337 2.6166666666666667-6.33 6.366666666666665-2.6549999999999994 6.403333333333334 2.6549999999999994l2.1133333333333333 1.8783333333333339-2.583333333333334 2.1883333333333344-1.875-1.6416666666666657q-1.7166666666666668-1.7166666666666668-4.063333333333334-1.7166666666666668t-3.9866666666666664 1.6383333333333336-1.6400000000000001 3.9833333333333343 1.6400000000000001 3.986666666666668 3.9833333333333343 1.6400000000000006 3.9866666666666664-1.6400000000000006q5.546666666666667-4.843333333333334 7.033333333333331-6.25l4.686666666666667-4.140000000000001q2.585000000000001-2.5766666666666698 6.335000000000001-2.5766666666666698z' })
                )
            );
        }
    }]);

    return MdAllInclusive;
}(React.Component);

exports.default = MdAllInclusive;
module.exports = exports['default'];