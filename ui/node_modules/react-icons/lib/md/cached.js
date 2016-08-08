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

var MdCached = function (_React$Component) {
    _inherits(MdCached, _React$Component);

    function MdCached() {
        _classCallCheck(this, MdCached);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdCached).apply(this, arguments));
    }

    _createClass(MdCached, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm10 20h5l-6.640000000000001 6.640000000000001-6.716666666666666-6.640000000000001h5q0-5.466666666666667 3.9433333333333342-9.413333333333334t9.413333333333332-3.9449999999999994q3.9066666666666663 0 7.109999999999999 2.1100000000000003l-2.421666666666667 2.423333333333334q-2.109999999999996-1.1750000000000007-4.688333333333333-1.1750000000000007-4.139999999999999 0-7.07 2.9333333333333336t-2.9299999999999997 7.066666666666666z m21.64-6.640000000000001l6.716666666666669 6.640000000000001h-5q0 5.466666666666669-3.943333333333335 9.413333333333334t-9.413333333333334 3.9450000000000003q-3.9066666666666663 0-7.109999999999999-2.1099999999999994l2.421666666666667-2.423333333333332q2.1099999999999994 1.1749999999999972 4.688333333333333 1.1749999999999972 4.140000000000001 0 7.07-2.9333333333333336t2.9299999999999997-7.066666666666666h-5z' })
                )
            );
        }
    }]);

    return MdCached;
}(React.Component);

exports.default = MdCached;
module.exports = exports['default'];