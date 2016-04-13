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

var MdHttp = function (_React$Component) {
    _inherits(MdHttp, _React$Component);

    function MdHttp() {
        _classCallCheck(this, MdHttp);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdHttp).apply(this, arguments));
    }

    _createClass(MdHttp, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm35.86 19.14v-1.6400000000000006h-3.3599999999999994v1.6400000000000006h3.3599999999999994z m0-4.140000000000001q1.0166666666666657 0 1.7583333333333329 0.7416666666666671t0.7433333333333323 1.7583333333333329v1.6400000000000006q0 1.0166666666666657-0.7416666666666671 1.7583333333333329t-1.7566666666666677 0.7399999999999984h-3.3633333333333297v3.361666666666668h-2.5v-10h5.859999999999999z m-15 2.5v-2.5h7.5v2.5h-2.5v7.5h-2.5v-7.5h-2.5z m-9.22 0v-2.5h7.500000000000002v2.5h-2.5v7.5h-2.5v-7.5h-2.5z m-4.140000000000001 0.8599999999999994v-3.3599999999999994h2.5000000000000018v10h-2.5v-4.140000000000001h-3.3599999999999994v4.140000000000001h-2.5v-10h2.5v3.3599999999999994h3.3599999999999994z' })
                )
            );
        }
    }]);

    return MdHttp;
}(React.Component);

exports.default = MdHttp;
module.exports = exports['default'];