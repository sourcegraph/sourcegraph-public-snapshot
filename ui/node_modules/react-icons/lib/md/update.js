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

var MdUpdate = function (_React$Component) {
    _inherits(MdUpdate, _React$Component);

    function MdUpdate() {
        _classCallCheck(this, MdUpdate);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdUpdate).apply(this, arguments));
    }

    _createClass(MdUpdate, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm20.86 13.360000000000001v7.033333333333333l5.783333333333335 3.513333333333332-1.173333333333332 2.0333333333333314-7.109999999999999-4.300000000000001v-8.27833333333333h2.5z m14.14 3.516666666666664h-11.328333333333333l4.611666666666665-4.693333333333332q-3.4400000000000013-3.4383333333333344-8.243333333333332-3.4766666666666666t-8.239999999999998 3.323333333333334q-3.3633333333333333 3.4366666666666656-3.3633333333333333 8.125t3.3599999999999994 8.125 8.203333333333333 3.4366666666666674 8.283333333333331-3.4366666666666674q3.3599999999999994-3.3599999999999994 3.3599999999999994-8.125h3.3599999999999994q0 6.171666666666667-4.375 10.466666666666669-4.3749999999999964 4.37833333333333-10.624999999999996 4.37833333333333t-10.625-4.376666666666665q-4.375-4.296666666666667-4.375-10.43t4.375-10.510000000000002 10.546666666666667-4.373333333333334 10.546666666666667 4.374999999999999l4.533333333333335-4.684999999999999v11.875z' })
                )
            );
        }
    }]);

    return MdUpdate;
}(React.Component);

exports.default = MdUpdate;
module.exports = exports['default'];