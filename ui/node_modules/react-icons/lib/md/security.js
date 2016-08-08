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

var MdSecurity = function (_React$Component) {
    _inherits(MdSecurity, _React$Component);

    function MdSecurity() {
        _classCallCheck(this, MdSecurity);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdSecurity).apply(this, arguments));
    }

    _createClass(MdSecurity, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm20 1.6400000000000001l15 6.716666666666667v10.000000000000002q0 6.954999999999998-4.296666666666667 12.696666666666665t-10.703333333333333 7.305q-6.406666666666666-1.5633333333333326-10.703333333333333-7.305t-4.296666666666667-12.695v-10z m0 18.36v14.921666666666667q4.609999999999999-1.4833333333333343 7.813333333333333-5.586666666666666t3.828333333333333-9.333333333333336h-11.641666666666666z m0 0v-14.688333333333333l-11.639999999999999 5.154999999999999v9.533333333333333h11.639999999999999z' })
                )
            );
        }
    }]);

    return MdSecurity;
}(React.Component);

exports.default = MdSecurity;
module.exports = exports['default'];