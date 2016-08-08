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

var TiUser = function (_React$Component) {
    _inherits(TiUser, _React$Component);

    function TiUser() {
        _classCallCheck(this, TiUser);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiUser).apply(this, arguments));
    }

    _createClass(TiUser, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm28.333333333333336 15c0-2.3000000000000007-0.9333333333333336-4.383333333333333-2.4400000000000013-5.8916666666666675s-3.5933333333333337-2.4416666666666655-5.893333333333334-2.4416666666666655-4.383333333333333 0.9333333333333336-5.893333333333334 2.4416666666666673c-1.506666666666666 1.5083333333333329-2.4399999999999977 3.591666666666667-2.4399999999999977 5.891666666666666s0.9333333333333336 4.383333333333333 2.4399999999999995 5.891666666666666c1.509999999999998 1.5083333333333329 3.593333333333332 2.44166666666667 5.893333333333333 2.44166666666667s4.383333333333333-0.9333333333333336 5.893333333333334-2.4416666666666664c1.5066666666666677-1.5083333333333329 2.4400000000000013-3.5916666666666686 2.4400000000000013-5.891666666666669z m-18.333333333333336 16.666666666666668c0 1.6666666666666679 3.75 3.333333333333332 10 3.333333333333332 5.863333333333333 0 10-1.6666666666666643 10-3.333333333333332 0-3.333333333333332-3.923333333333332-6.666666666666668-10-6.666666666666668-6.25 0-10 3.333333333333332-10 6.666666666666668z' })
                )
            );
        }
    }]);

    return TiUser;
}(React.Component);

exports.default = TiUser;
module.exports = exports['default'];