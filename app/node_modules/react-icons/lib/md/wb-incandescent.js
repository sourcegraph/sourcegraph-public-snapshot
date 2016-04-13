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

var MdWbIncandescent = function (_React$Component) {
    _inherits(MdWbIncandescent, _React$Component);

    function MdWbIncandescent() {
        _classCallCheck(this, MdWbIncandescent);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdWbIncandescent).apply(this, arguments));
    }

    _createClass(MdWbIncandescent, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm28.75 30.233333333333334l2.3433333333333337-2.2633333333333354 2.9666666666666686 2.9666666666666686-2.3416666666666686 2.344999999999999z m4.609999999999999-12.733333333333334h5v3.3599999999999994h-5v-3.3599999999999994z m-8.36-6.953333333333333q2.2666666666666657 1.3283333333333331 3.633333333333333 3.5933333333333337t1.3666666666666671 5q0 4.140000000000001-2.9299999999999997 7.07t-7.07 2.9300000000000033-7.07-2.9299999999999997-2.9299999999999997-7.070000000000004q0-2.7333333333333343 1.3666666666666671-5t3.633333333333333-3.5933333333333337v-8.046666666666667h10v8.046666666666667z m-18.36 6.953333333333333v3.3599999999999994h-5v-3.3599999999999994h5z m11.719999999999999 19.921666666666667v-4.921666666666667h3.2833333333333314v4.921666666666667h-3.2833333333333314z m-12.421666666666667-6.483333333333334l2.966666666666667-3.046666666666667 2.3450000000000006 2.3416666666666686-2.966666666666667 3.049999999999997z' })
                )
            );
        }
    }]);

    return MdWbIncandescent;
}(React.Component);

exports.default = MdWbIncandescent;
module.exports = exports['default'];