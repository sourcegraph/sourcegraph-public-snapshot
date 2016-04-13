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

var MdLocalRestaurant = function (_React$Component) {
    _inherits(MdLocalRestaurant, _React$Component);

    function MdLocalRestaurant() {
        _classCallCheck(this, MdLocalRestaurant);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdLocalRestaurant).apply(this, arguments));
    }

    _createClass(MdLocalRestaurant, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm24.766666666666666 19.216666666666665l-2.423333333333332 2.423333333333332 11.483333333333334 11.48333333333333-2.3416666666666686 2.344999999999999-11.485-11.484999999999992-11.483333333333333 11.483333333333334-2.3450000000000006-2.3416666666666686 16.25-16.25q-0.9383333333333326-1.875-0.2716666666666683-4.375t2.616666666666667-4.453333333333333q2.4200000000000017-2.421666666666667 5.388333333333335-2.7733333333333334t4.766666666666666 1.4450000000000003 1.4433333333333351 4.806666666666667-2.7749999999999986 5.4300000000000015q-1.9533333333333331 1.9533333333333331-4.453333333333333 2.578333333333333t-4.375-0.31666666666666643z m-11.25 3.0500000000000007l-7.033333333333332-7.033333333333331q-1.9533333333333331-1.9533333333333331-1.9533333333333331-4.688333333333334t1.9533333333333331-4.6883333333333335l11.716666666666665 11.64z' })
                )
            );
        }
    }]);

    return MdLocalRestaurant;
}(React.Component);

exports.default = MdLocalRestaurant;
module.exports = exports['default'];