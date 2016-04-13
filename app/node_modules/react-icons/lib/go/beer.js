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

var GoBeer = function (_React$Component) {
    _inherits(GoBeer, _React$Component);

    function GoBeer() {
        _classCallCheck(this, GoBeer);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(GoBeer).apply(this, arguments));
    }

    _createClass(GoBeer, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm35 10h-7.5v-5c0-2.7750000000000004-6.172499999999999-5-13.75-5s-13.75 2.225-13.75 5v30c0 2.7749999999999986 6.172499999999999 5 13.75 5s13.75-2.2250000000000014 13.75-5v-5h7.5s2.5-1.1724999999999994 2.5-2.5v-15s-1.1325000000000003-2.5-2.5-2.5z m-27.5 22.5h-2.5v-20h2.5v20z m7.5 2.5h-2.5v-20h2.5v20z m7.5-2.5h-2.5v-20h2.5v20z m-8.75-25c-4.84375 0-8.75-1.1325000000000003-8.75-2.5s3.90625-2.5 8.75-2.5 8.75 1.1325000000000003 8.75 2.5-3.90625 2.5-8.75 2.5z m18.75 17.5h-5v-10h5v10z' })
                )
            );
        }
    }]);

    return GoBeer;
}(React.Component);

exports.default = GoBeer;
module.exports = exports['default'];