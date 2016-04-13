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

var FaNewspaperO = function (_React$Component) {
    _inherits(FaNewspaperO, _React$Component);

    function FaNewspaperO() {
        _classCallCheck(this, FaNewspaperO);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaNewspaperO).apply(this, arguments));
    }

    _createClass(FaNewspaperO, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm20 12.5h-7.5v7.5h7.5v-7.5z m2.5 12.5v2.5h-12.5v-2.5h12.5z m0-15v12.5h-12.5v-12.5h12.5z m12.5 15v2.5h-10v-2.5h10z m0-5v2.5h-10v-2.5h10z m0-5v2.5h-10v-2.5h10z m0-5v2.5h-10v-2.5h10z m-30 18.75v-18.75h-2.5v18.75q0 0.5075000000000003 0.37124999999999986 0.8787500000000001t0.8787500000000001 0.37124999999999986 0.8787500000000001-0.37124999999999986 0.37124999999999986-0.8787500000000001z m32.5 0v-21.25h-30v21.25q0 0.6449999999999996-0.21499999999999986 1.25h28.965q0.5075000000000003 0 0.8787499999999966-0.37124999999999986t0.3712500000000034-0.8787500000000001z m2.5-23.75v23.75q0 1.5625-1.09375 2.65625t-2.65625 1.09375h-32.5q-1.5625 0-2.65625-1.09375t-1.09375-2.65625v-21.25h5v-2.5h35z' })
                )
            );
        }
    }]);

    return FaNewspaperO;
}(React.Component);

exports.default = FaNewspaperO;
module.exports = exports['default'];