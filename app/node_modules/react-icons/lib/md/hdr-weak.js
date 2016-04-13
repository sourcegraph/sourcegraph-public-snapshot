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

var MdHdrWeak = function (_React$Component) {
    _inherits(MdHdrWeak, _React$Component);

    function MdHdrWeak() {
        _classCallCheck(this, MdHdrWeak);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdHdrWeak).apply(this, arguments));
    }

    _createClass(MdHdrWeak, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm28.36 26.64q2.7333333333333343 0 4.688333333333333-1.9533333333333331t1.951666666666668-4.686666666666667-1.9500000000000028-4.690000000000001-4.690000000000001-1.9533333333333331-4.726666666666667 1.9533333333333331-1.9916666666666636 4.690000000000001 1.9916666666666671 4.686666666666667 4.726666666666667 1.9533333333333331z m0-16.64q4.140000000000001 0 7.07 2.9299999999999997t2.9299999999999997 7.07-2.9299999999999997 7.07-7.07 2.9299999999999997-7.07-2.9299999999999997-2.9299999999999997-7.07 2.9299999999999997-7.07 7.07-2.9299999999999997z m-20 3.3599999999999994q2.7333333333333343 0 4.688333333333333 1.9533333333333331t1.951666666666668 4.686666666666667-1.9499999999999993 4.690000000000001-4.690000000000001 1.9533333333333331-4.726666666666667-1.9533333333333331-1.9916666666666663-4.690000000000001 1.9916666666666671-4.683333333333334 4.726666666666666-1.9533333333333331z' })
                )
            );
        }
    }]);

    return MdHdrWeak;
}(React.Component);

exports.default = MdHdrWeak;
module.exports = exports['default'];