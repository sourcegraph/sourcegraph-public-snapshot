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

var MdFolderShared = function (_React$Component) {
    _inherits(MdFolderShared, _React$Component);

    function MdFolderShared() {
        _classCallCheck(this, MdFolderShared);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdFolderShared).apply(this, arguments));
    }

    _createClass(MdFolderShared, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm31.640000000000004 28.36v-1.7166666666666686q0-1.4866666666666681-2.3049999999999997-2.383333333333333t-4.333333333333332-0.8999999999999986-4.338333333333335 0.8999999999999986-2.3049999999999997 2.383333333333333v1.7166666666666686h13.283333333333335z m-6.640000000000004-13.36q-1.3283333333333331 0-2.3433333333333337 1.0166666666666657t-1.0166666666666657 2.3416666666666686 1.0166666666666657 2.3049999999999997 2.3433333333333337 0.9766666666666666 2.3433333333333337-0.9766666666666666 1.0166666666666657-2.3049999999999997-1.0166666666666657-2.3433333333333337-2.3433333333333337-1.0150000000000006z m8.36-5q1.3283333333333331 0 2.3049999999999997 1.0166666666666657t0.9750000000000014 2.341666666666667v16.641666666666666q0 1.326666666666668-0.9766666666666666 2.3416666666666686t-2.306666666666665 1.0166666666666657h-26.713333333333335q-1.330000000000001 0-2.3066666666666675-1.0166666666666657t-0.9766666666666666-2.3416666666666686v-20q0-1.33 0.9766666666666666-2.3450000000000006t2.3050000000000006-1.0166666666666666h9.999999999999998l3.3583333333333343 3.3616666666666672h13.361666666666665z' })
                )
            );
        }
    }]);

    return MdFolderShared;
}(React.Component);

exports.default = MdFolderShared;
module.exports = exports['default'];