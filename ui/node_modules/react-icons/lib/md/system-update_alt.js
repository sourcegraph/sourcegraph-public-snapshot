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

var MdSystemUpdateAlt = function (_React$Component) {
    _inherits(MdSystemUpdateAlt, _React$Component);

    function MdSystemUpdateAlt() {
        _classCallCheck(this, MdSystemUpdateAlt);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdSystemUpdateAlt).apply(this, arguments));
    }

    _createClass(MdSystemUpdateAlt, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm35 5.86q1.3283333333333331 0 2.3433333333333337 0.9766666666666666t1.0166666666666657 2.3050000000000006v23.358333333333334q0 1.3299999999999983-1.0166666666666657 2.344999999999999t-2.3433333333333337 1.0166666666666657h-30q-1.3283333333333331 0-2.3433333333333333-1.0166666666666657t-1.0166666666666666-2.344999999999999v-23.35666666666667q0-1.3283333333333314 1.0166666666666666-2.304999999999998t2.3433333333333333-0.9783333333333326h10v3.283333333333334h-10v23.356666666666666h30v-23.35666666666667h-10v-3.2833333333333314h10z m-15 21.64l-6.640000000000001-6.640000000000001h5v-15h3.2833333333333314v15h5z' })
                )
            );
        }
    }]);

    return MdSystemUpdateAlt;
}(React.Component);

exports.default = MdSystemUpdateAlt;
module.exports = exports['default'];