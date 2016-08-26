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

var MdPersonAdd = function (_React$Component) {
    _inherits(MdPersonAdd, _React$Component);

    function MdPersonAdd() {
        _classCallCheck(this, MdPersonAdd);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdPersonAdd).apply(this, arguments));
    }

    _createClass(MdPersonAdd, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm25 23.4c4.5 0 13.4 2.1 13.4 6.6v3.4h-26.8v-3.4c0-4.5 8.9-6.6 13.4-6.6z m-15-6.8h5v3.4h-5v5h-3.4v-5h-5v-3.4h5v-5h3.4v5z m15 3.4c-3.7 0-6.6-3-6.6-6.6s2.9-6.8 6.6-6.8 6.6 3.1 6.6 6.8-2.9 6.6-6.6 6.6z' })
                )
            );
        }
    }]);

    return MdPersonAdd;
}(React.Component);

exports.default = MdPersonAdd;
module.exports = exports['default'];