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

var MdShopTwo = function (_React$Component) {
    _inherits(MdShopTwo, _React$Component);

    function MdShopTwo() {
        _classCallCheck(this, MdShopTwo);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdShopTwo).apply(this, arguments));
    }

    _createClass(MdShopTwo, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm20 25l9.14-6.640000000000001-9.14-5v11.64z m0-20v3.3599999999999994h6.640000000000001v-3.3599999999999994h-6.640000000000001z m10 3.3599999999999994h8.36v18.283333333333335q0 1.4050000000000011-0.9766666666666666 2.383333333333333t-2.383333333333333 0.9733333333333327h-23.36q-1.4066666666666663 0-2.3433333333333337-0.9750000000000014t-0.9383333333333326-2.383333333333333v-18.28333333333333h8.283333333333331v-3.3583333333333343q0-1.4066666666666667 0.9750000000000014-2.3833333333333333t2.383333333333333-0.9766666666666666h6.640000000000001q1.4066666666666663 0 2.383333333333333 0.9766666666666666t0.9766666666666666 2.3833333333333333v3.3599999999999994z m-25 6.640000000000001v18.36h26.64q0 1.4066666666666663-0.9383333333333326 2.3433333333333337t-2.3416666666666686 0.9383333333333326h-23.36q-1.4083333333333332 0-2.3833333333333333-0.9383333333333326t-0.9783333333333333-2.3433333333333337v-18.36h3.3616666666666664z' })
                )
            );
        }
    }]);

    return MdShopTwo;
}(React.Component);

exports.default = MdShopTwo;
module.exports = exports['default'];