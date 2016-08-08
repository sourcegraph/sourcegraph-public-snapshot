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

var MdPolymer = function (_React$Component) {
    _inherits(MdPolymer, _React$Component);

    function MdPolymer() {
        _classCallCheck(this, MdPolymer);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdPolymer).apply(this, arguments));
    }

    _createClass(MdPolymer, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm31.640000000000004 6.640000000000001l7.4999999999999964 13.36-7.5 13.36h-6.640000000000001l7.5-13.36-4.375-7.733333333333334-13.125 21.090000000000003h-6.639999999999999l-7.500000000000001-13.35666666666667 7.499999999999999-13.361666666666666h6.640000000000001l-7.5 13.361666666666666 4.375 7.733333333333334 13.125-21.093333333333334h6.640000000000001z' })
                )
            );
        }
    }]);

    return MdPolymer;
}(React.Component);

exports.default = MdPolymer;
module.exports = exports['default'];