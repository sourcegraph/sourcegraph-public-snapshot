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

var MdTimer = function (_React$Component) {
    _inherits(MdTimer, _React$Component);

    function MdTimer() {
        _classCallCheck(this, MdTimer);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdTimer).apply(this, arguments));
    }

    _createClass(MdTimer, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm20 33.36q4.843333333333334 0 8.241666666666667-3.4383333333333326t3.3999999999999986-8.283333333333331-3.3999999999999986-8.24-8.241666666666667-3.398333333333335-8.241666666666667 3.3999999999999986-3.4000000000000004 8.240000000000002 3.4000000000000004 8.283333333333331 8.241666666666667 3.4350000000000023z m11.716666666666669-21.016666666666666q3.2833333333333314 4.140000000000001 3.2833333333333314 9.296666666666667 0 6.171666666666667-4.375 10.586666666666666t-10.625 4.413333333333341-10.625-4.413333333333334-4.375-10.586666666666673 4.375-10.586666666666668 10.625-4.413333333333332q5.078333333333333 0 9.375 3.3599999999999994l2.3433333333333337-2.421666666666667q1.25 1.0166666666666657 2.344999999999999 2.3433333333333337z m-13.356666666666666 11.016666666666666v-10h3.283333333333335v10h-3.2833333333333314z m6.639999999999997-21.72v3.3599999999999994h-10v-3.36h10z' })
                )
            );
        }
    }]);

    return MdTimer;
}(React.Component);

exports.default = MdTimer;
module.exports = exports['default'];