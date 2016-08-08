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

var TiHome = function (_React$Component) {
    _inherits(TiHome, _React$Component);

    function TiHome() {
        _classCallCheck(this, TiHome);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiHome).apply(this, arguments));
    }

    _createClass(TiHome, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm20 5s-10.31 8.9-16.07166666666667 13.719999999999999c-0.33833333333333115 0.3066666666666684-0.5949999999999975 0.7533333333333339-0.5949999999999975 1.2800000000000011 0 0.9216666666666669 0.7449999999999997 1.6666666666666679 1.6666666666666665 1.6666666666666679h3.333333333333334v11.666666666666668c0 0.9216666666666669 0.7449999999999992 1.6666666666666643 1.666666666666666 1.6666666666666643h5c0.9216666666666669 0 1.6666666666666679-0.7466666666666697 1.6666666666666679-1.6666666666666643v-6.666666666666668h6.666666666666668v6.666666666666668c0 0.9200000000000017 0.745000000000001 1.6666666666666643 1.6666666666666679 1.6666666666666643h5c0.9216666666666669 0 1.6666666666666679-0.7449999999999974 1.6666666666666679-1.6666666666666643v-11.666666666666668h3.3333333333333357c0.9216666666666669 0 1.6666666666666643-0.745000000000001 1.6666666666666643-1.6666666666666679 0-0.5266666666666673-0.2566666666666677-0.9733333333333327-0.6383333333333354-1.2800000000000011-5.721666666666668-4.8199999999999985-16.028333333333336-13.719999999999999-16.028333333333336-13.719999999999999z' })
                )
            );
        }
    }]);

    return TiHome;
}(React.Component);

exports.default = TiHome;
module.exports = exports['default'];