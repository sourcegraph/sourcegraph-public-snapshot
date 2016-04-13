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

var TiCancel = function (_React$Component) {
    _inherits(TiCancel, _React$Component);

    function TiCancel() {
        _classCallCheck(this, TiCancel);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiCancel).apply(this, arguments));
    }

    _createClass(TiCancel, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm20 6.666666666666667c-7.350000000000001 0-13.333333333333334 5.983333333333333-13.333333333333334 13.333333333333332s5.9833333333333325 13.333333333333336 13.333333333333334 13.333333333333336 13.333333333333336-5.983333333333334 13.333333333333336-13.333333333333336-5.983333333333334-13.333333333333334-13.333333333333336-13.333333333333334z m-8.333333333333334 13.333333333333332c0-1.3866666666666667 0.3733333333333331-2.673333333333332 0.9733333333333327-3.8249999999999993l11.183333333333332 11.183333333333334c-1.1499999999999986 0.6000000000000014-2.4366666666666674 0.9750000000000014-3.823333333333334 0.9750000000000014-4.595000000000001 0-8.333333333333334-3.7383333333333333-8.333333333333334-8.333333333333336z m15.693333333333333 3.8249999999999993l-11.183333333333334-11.183333333333334c1.1483333333333334-0.6016666666666648 2.4350000000000023-0.9749999999999979 3.823333333333334-0.9749999999999979 4.594999999999999 0 8.333333333333336 3.7383333333333333 8.333333333333336 8.333333333333332 0 1.3866666666666667-0.37333333333333485 2.673333333333332-0.9733333333333327 3.8249999999999993z' })
                )
            );
        }
    }]);

    return TiCancel;
}(React.Component);

exports.default = TiCancel;
module.exports = exports['default'];