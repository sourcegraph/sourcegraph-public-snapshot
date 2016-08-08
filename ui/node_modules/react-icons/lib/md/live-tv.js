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

var MdLiveTv = function (_React$Component) {
    _inherits(MdLiveTv, _React$Component);

    function MdLiveTv() {
        _classCallCheck(this, MdLiveTv);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdLiveTv).apply(this, arguments));
    }

    _createClass(MdLiveTv, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm15 16.64l11.64 6.716666666666669-11.64 6.643333333333331v-13.363333333333333z m20 16.72v-20h-30v20h30z m0-23.36q1.3283333333333331 0 2.3433333333333337 0.9766666666666666t1.0166666666666657 2.383333333333333v20q0 1.3283333333333331-1.0166666666666657 2.3049999999999997t-2.3433333333333337 0.9750000000000085h-30q-1.3283333333333331 0-2.3433333333333333-0.9766666666666666t-1.0166666666666666-2.306666666666665v-20q0-1.4066666666666663 1.0166666666666666-2.383333333333333t2.3433333333333333-0.9733333333333434h12.656666666666666l-5.466666666666667-5.466666666666667 1.17-1.1733333333333333 6.640000000000001 6.640000000000001 6.640000000000001-6.640000000000001 1.1716666666666669 1.1716666666666669-5.466666666666665 5.468333333333334h12.654999999999998z' })
                )
            );
        }
    }]);

    return MdLiveTv;
}(React.Component);

exports.default = MdLiveTv;
module.exports = exports['default'];