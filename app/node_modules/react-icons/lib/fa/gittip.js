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

var FaGittip = function (_React$Component) {
    _inherits(FaGittip, _React$Component);

    function FaGittip() {
        _classCallCheck(this, FaGittip);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaGittip).apply(this, arguments));
    }

    _createClass(FaGittip, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm20.111428571428572 29.062857142857144l7.814285714285713-10.557142857142857q0.35714285714285765-0.49285714285714377 0.5457142857142863-1.3185714285714276t-0.13428571428571345-1.895714285714286-1.37142857142857-1.7628571428571433q-0.894285714285715-0.581428571428571-1.8542857142857159-0.5714285714285712t-1.6400000000000006 0.39142857142857146-1.217142857142857 1.0042857142857144q-0.8028571428571425 0.8928571428571423-2.1428571428571423 0.8928571428571423-1.3171428571428585 0-2.120000000000001-0.8928571428571423-0.5357142857142847-0.6257142857142863-1.217142857142857-1.0042857142857144t-1.6400000000000006-0.39142857142857324-1.8757142857142863 0.5714285714285712q-1.025714285714285 0.6899999999999995-1.3485714285714288 1.7614285714285707t-0.13428571428571345 1.8971428571428568 0.5471428571428572 1.3185714285714276z m17.031428571428574-9.062857142857144q0 4.665714285714287-2.299999999999997 8.604285714285716t-6.237142857142857 6.238571428571426-8.605714285714292 2.3000000000000043-8.6-2.3000000000000043-6.242857142857143-6.238571428571426-2.295714285714286-8.604285714285716 2.3000000000000003-8.604285714285714 6.234285714285714-6.238571428571428 8.604285714285714-2.3000000000000003 8.605714285714285 2.3000000000000003 6.238571428571426 6.238571428571428 2.298571428571435 8.604285714285714z' })
                )
            );
        }
    }]);

    return FaGittip;
}(React.Component);

exports.default = FaGittip;
module.exports = exports['default'];