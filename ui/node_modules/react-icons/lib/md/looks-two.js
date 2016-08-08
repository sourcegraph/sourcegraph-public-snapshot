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

var MdLooksTwo = function (_React$Component) {
    _inherits(MdLooksTwo, _React$Component);

    function MdLooksTwo() {
        _classCallCheck(this, MdLooksTwo);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdLooksTwo).apply(this, arguments));
    }

    _createClass(MdLooksTwo, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm25 18.36v-3.3599999999999994q0-1.4066666666666663-1.0166666666666657-2.383333333333333t-2.3416666666666686-0.9766666666666666h-6.641666666666666v3.3599999999999994h6.641666666666666v3.3599999999999994h-3.2833333333333314q-1.4050000000000011 0-2.383333333333333 0.9383333333333326t-0.9750000000000014 2.3416666666666686v6.716666666666669h10v-3.356666666666669h-6.640000000000001v-3.3599999999999994h3.2833333333333314q1.4050000000000011 0 2.383333333333333-0.9383333333333326t0.9733333333333363-2.3416666666666686z m6.640000000000004-13.36q1.3283333333333367 0 2.34333333333333 1.0166666666666666t1.0166666666666657 2.3400000000000007v23.28333333333333q0 1.326666666666668-1.0166666666666657 2.3416666666666686t-2.3433333333333337 1.0166666666666657h-23.28333333333333q-1.3266666666666689 0-2.3416666666666686-1.0166666666666657t-1.0150000000000006-2.341666666666665v-23.28333333333334q0-1.3266666666666653 1.0166666666666666-2.341666666666665t2.3400000000000007-1.0150000000000006h23.28333333333333z' })
                )
            );
        }
    }]);

    return MdLooksTwo;
}(React.Component);

exports.default = MdLooksTwo;
module.exports = exports['default'];