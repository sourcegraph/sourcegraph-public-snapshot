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

var MdRepeatOne = function (_React$Component) {
    _inherits(MdRepeatOne, _React$Component);

    function MdRepeatOne() {
        _classCallCheck(this, MdRepeatOne);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdRepeatOne).apply(this, arguments));
    }

    _createClass(MdRepeatOne, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm21.64 25h-2.5v-6.640000000000001h-2.5v-1.7166666666666686l3.3599999999999994-1.6433333333333309h1.6400000000000006v10z m6.719999999999999 3.3599999999999994v-6.716666666666669h3.2833333333333314v10h-20v5l-6.643333333333331-6.643333333333331 6.641666666666666-6.638333333333335v5h16.716666666666665z m-16.72-16.72v6.716666666666669h-3.283333333333333v-10h20v-5l6.643333333333331 6.643333333333331-6.641666666666666 6.638333333333335v-5h-16.71666666666667z' })
                )
            );
        }
    }]);

    return MdRepeatOne;
}(React.Component);

exports.default = MdRepeatOne;
module.exports = exports['default'];