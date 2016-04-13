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

var FaBed = function (_React$Component) {
    _inherits(FaBed, _React$Component);

    function FaBed() {
        _classCallCheck(this, FaBed);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaBed).apply(this, arguments));
    }

    _createClass(FaBed, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm5 22.5h33.75q0.5075000000000003 0 0.8787499999999966 0.37124999999999986t0.3712500000000034 0.8787500000000001v8.75h-5v-5h-30v5h-5v-23.75q0-0.5075000000000003 0.37124999999999997-0.8787500000000001t0.87875-0.37124999999999986h2.5q0.5075000000000003 0 0.8787500000000001 0.37124999999999986t0.37124999999999986 0.8787500000000001v13.75z m11.25-6.25q0-2.0700000000000003-1.4649999999999999-3.535t-3.535-1.4649999999999999-3.535 1.4649999999999999-1.4649999999999999 3.535 1.4649999999999999 3.535 3.535 1.4649999999999999 3.535-1.4649999999999999 1.4649999999999999-3.535z m23.75 5v-1.25q0-3.1050000000000004-2.197499999999998-5.3025t-5.302500000000002-2.1975h-13.75q-0.5075000000000003 0-0.8787500000000001 0.37124999999999986t-0.37124999999999986 0.8787500000000001v7.5h22.5z' })
                )
            );
        }
    }]);

    return FaBed;
}(React.Component);

exports.default = FaBed;
module.exports = exports['default'];