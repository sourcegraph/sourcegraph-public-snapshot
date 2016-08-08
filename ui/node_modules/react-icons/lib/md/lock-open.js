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

var MdLockOpen = function (_React$Component) {
    _inherits(MdLockOpen, _React$Component);

    function MdLockOpen() {
        _classCallCheck(this, MdLockOpen);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdLockOpen).apply(this, arguments));
    }

    _createClass(MdLockOpen, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm30 33.36v-16.716666666666665h-20v16.716666666666665h20z m0-20q1.3283333333333331 0 2.3433333333333337 0.9766666666666666t1.0166666666666657 2.3049999999999997v16.71666666666667q0 1.3299999999999983-1.0166666666666657 2.306666666666665t-2.3433333333333337 0.9750000000000085h-20q-1.3283333333333331 0-2.3433333333333337-0.9766666666666666t-1.0166666666666666-2.306666666666665v-16.713333333333342q0-1.33 1.0166666666666666-2.3066666666666666t2.3433333333333337-0.9766666666666666h15.156666666666666v-3.360000000000001q0-2.1100000000000003-1.5233333333333334-3.6333333333333337t-3.633333333333333-1.5233333333333325-3.633333333333333 1.5233333333333334-1.5233333333333317 3.633333333333333h-3.203333333333335q0-3.4383333333333344 2.461666666666666-5.9t5.898333333333333-2.458333333333333 5.899999999999999 2.461666666666667 2.4583333333333357 5.8966666666666665v3.3599999999999994h1.6416666666666657z m-10 15q-1.3283333333333331 0-2.3433333333333337-1.0166666666666657t-1.0166666666666657-2.3416666666666686 1.0166666666666657-2.3416666666666686 2.3433333333333337-1.0166666666666657 2.3433333333333337 1.0166666666666657 1.0166666666666657 2.3433333333333337-1.0166666666666657 2.344999999999999-2.3433333333333337 1.0166666666666657z' })
                )
            );
        }
    }]);

    return MdLockOpen;
}(React.Component);

exports.default = MdLockOpen;
module.exports = exports['default'];