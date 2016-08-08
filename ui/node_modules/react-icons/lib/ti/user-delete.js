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

var TiUserDelete = function (_React$Component) {
    _inherits(TiUserDelete, _React$Component);

    function TiUserDelete() {
        _classCallCheck(this, TiUserDelete);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiUserDelete).apply(this, arguments));
    }

    _createClass(TiUserDelete, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm35 23.333333333333336h-10c-0.9216666666666669 0-1.6666666666666679-0.745000000000001-1.6666666666666679-1.6666666666666679s0.745000000000001-1.6666666666666679 1.6666666666666679-1.6666666666666679h10c0.9216666666666669 0 1.6666666666666643 0.745000000000001 1.6666666666666643 1.6666666666666679s-0.7449999999999974 1.6666666666666679-1.6666666666666643 1.6666666666666679z m-11.666666666666664-8.333333333333336c0 2.3000000000000007-0.9333333333333336 4.383333333333333-2.4400000000000013 5.891666666666666s-3.5933333333333337 2.44166666666667-5.893333333333334 2.44166666666667-4.383333333333333-0.9333333333333336-5.893333333333334-2.4416666666666664c-1.506666666666666-1.5083333333333329-2.4399999999999986-3.5916666666666686-2.4399999999999986-5.891666666666669s0.9333333333333336-4.383333333333333 2.4400000000000004-5.8916666666666675c1.5099999999999998-1.5083333333333329 3.5933333333333337-2.4416666666666655 5.893333333333333-2.4416666666666655s4.383333333333333 0.9333333333333336 5.893333333333334 2.4416666666666673c1.5066666666666642 1.5083333333333329 2.4400000000000013 3.591666666666667 2.4400000000000013 5.891666666666666z m-8.333333333333336 10c-6.25 0-10 3.333333333333332-10 6.666666666666668 0 1.6666666666666679 3.75 3.333333333333332 10 3.333333333333332 5.863333333333333 0 10-1.6666666666666643 10-3.333333333333332s-3.923333333333332-6.666666666666668-10-6.666666666666668z' })
                )
            );
        }
    }]);

    return TiUserDelete;
}(React.Component);

exports.default = TiUserDelete;
module.exports = exports['default'];