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

var MdPhonelinkErase = function (_React$Component) {
    _inherits(MdPhonelinkErase, _React$Component);

    function MdPhonelinkErase() {
        _classCallCheck(this, MdPhonelinkErase);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdPhonelinkErase).apply(this, arguments));
    }

    _createClass(MdPhonelinkErase, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm31.640000000000004 1.6400000000000001q1.3283333333333367 0 2.34333333333333 1.0166666666666666t1.0166666666666657 2.3400000000000003v30.000000000000004q0 1.326666666666668-1.0166666666666657 2.3416666666666686t-2.3433333333333337 1.0166666666666657h-16.64q-1.3283333333333331 0-2.3433333333333337-1.0166666666666657t-1.0166666666666657-2.3383333333333383v-5h3.3616666666666664v3.3583333333333343h16.641666666666666v-26.71666666666667h-16.64333333333333v3.3583333333333343h-3.3566666666666674v-5q0-1.3283333333333331 1.0166666666666675-2.3433333333333333t2.34-1.0166666666666666h16.64z m-10 12.033333333333333l-6.640000000000004 6.6366666666666685 6.640000000000001 6.716666666666669-1.6400000000000006 1.6416666666666657-6.640000000000001-6.640000000000001-6.716666666666668 6.640000000000001-1.6433333333333318-1.6383333333333354 6.641666666666666-6.716666666666669-6.641666666666666-6.643333333333333 1.6416666666666666-1.6400000000000006 6.716666666666668 6.638333333333334 6.641666666666666-6.635z' })
                )
            );
        }
    }]);

    return MdPhonelinkErase;
}(React.Component);

exports.default = MdPhonelinkErase;
module.exports = exports['default'];