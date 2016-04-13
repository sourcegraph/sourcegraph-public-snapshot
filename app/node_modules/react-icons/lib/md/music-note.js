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

var MdMusicNote = function (_React$Component) {
    _inherits(MdMusicNote, _React$Component);

    function MdMusicNote() {
        _classCallCheck(this, MdMusicNote);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdMusicNote).apply(this, arguments));
    }

    _createClass(MdMusicNote, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm20 5h10v6.640000000000001h-6.640000000000001v16.716666666666665q0 2.736666666666668-1.9916666666666671 4.690000000000001t-4.724999999999998 1.9533333333333331-4.688333333333333-1.9533333333333331-1.9550000000000018-4.688333333333333 1.9533333333333331-4.726666666666667 4.688333333333333-1.9933333333333323q1.6400000000000006 0 3.3599999999999994 0.9383333333333326v-17.576666666666668z' })
                )
            );
        }
    }]);

    return MdMusicNote;
}(React.Component);

exports.default = MdMusicNote;
module.exports = exports['default'];