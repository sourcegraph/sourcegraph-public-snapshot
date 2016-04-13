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

var MdGolfCourse = function (_React$Component) {
    _inherits(MdGolfCourse, _React$Component);

    function MdGolfCourse() {
        _classCallCheck(this, MdGolfCourse);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdGolfCourse).apply(this, arguments));
    }

    _createClass(MdGolfCourse, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm28.36 9.843333333333334l-10 5.156666666666666v15.078333333333333q3.5933333333333337 0.1566666666666663 5.938333333333333 1.0933333333333337t2.3416666666666686 2.1883333333333326q0 1.3283333333333331-2.9299999999999997 2.3049999999999997t-7.07 0.9750000000000014-7.07-0.9766666666666666-2.9299999999999997-2.306666666666665q0-1.9533333333333331 5-2.8900000000000006v2.8900000000000006h3.3599999999999994v-30z m1.6400000000000006 22.656666666666666q0-1.0933333333333337 0.7033333333333331-1.7966666666666669t1.7966666666666669-0.7033333333333331 1.7966666666666669 0.7033333333333331 0.7033333333333331 1.7966666666666669-0.7033333333333331 1.7966666666666669-1.7966666666666669 0.7033333333333331-1.7966666666666669-0.7033333333333331-0.7033333333333331-1.7966666666666669z' })
                )
            );
        }
    }]);

    return MdGolfCourse;
}(React.Component);

exports.default = MdGolfCourse;
module.exports = exports['default'];