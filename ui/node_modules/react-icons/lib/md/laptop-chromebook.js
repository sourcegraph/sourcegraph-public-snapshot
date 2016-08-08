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

var MdLaptopChromebook = function (_React$Component) {
    _inherits(MdLaptopChromebook, _React$Component);

    function MdLaptopChromebook() {
        _classCallCheck(this, MdLaptopChromebook);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdLaptopChromebook).apply(this, arguments));
    }

    _createClass(MdLaptopChromebook, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm33.36 25v-16.64h-26.716666666666665v16.64h26.716666666666665z m-10 5v-1.6400000000000006h-6.716666666666669v1.6400000000000006h6.716666666666669z m13.280000000000001 0h3.3599999999999994v3.3599999999999994h-40v-3.3599999999999994h3.3600000000000003v-25h33.28333333333333v25z' })
                )
            );
        }
    }]);

    return MdLaptopChromebook;
}(React.Component);

exports.default = MdLaptopChromebook;
module.exports = exports['default'];