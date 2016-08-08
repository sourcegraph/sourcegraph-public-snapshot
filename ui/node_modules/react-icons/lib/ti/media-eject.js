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

var TiMediaEject = function (_React$Component) {
    _inherits(TiMediaEject, _React$Component);

    function TiMediaEject() {
        _classCallCheck(this, TiMediaEject);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiMediaEject).apply(this, arguments));
    }

    _createClass(TiMediaEject, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm28.333333333333336 26.666666666666668h-16.666666666666668c-1.8399999999999999 0-3.333333333333334 1.4916666666666671-3.333333333333334 3.333333333333332s1.493333333333334 3.3333333333333357 3.333333333333334 3.3333333333333357h16.666666666666668c1.8399999999999999 0 3.333333333333332-1.4933333333333323 3.333333333333332-3.333333333333332 0-1.841666666666665-1.4933333333333323-3.333333333333332-3.333333333333332-3.333333333333332z m2.388333333333332-8.993333333333336c-4.288333333333334-4.399999999999997-10.721666666666668-11.006666666666664-10.721666666666668-11.006666666666664l-10.721666666666668 11.006666666666668c-0.5833333333333321 0.6050000000000004-0.9449999999999985 1.4216666666666669-0.9449999999999985 2.3266666666666644 0 1.8399999999999999 1.493333333333334 3.333333333333332 3.333333333333334 3.333333333333332h16.666666666666668c1.8399999999999999 0 3.333333333333332-1.4933333333333323 3.333333333333332-3.333333333333332 0-0.9050000000000011-0.3633333333333333-1.7216666666666676-0.9450000000000003-2.326666666666668z' })
                )
            );
        }
    }]);

    return TiMediaEject;
}(React.Component);

exports.default = TiMediaEject;
module.exports = exports['default'];