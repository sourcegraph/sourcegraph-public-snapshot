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

var MdCake = function (_React$Component) {
    _inherits(MdCake, _React$Component);

    function MdCake() {
        _classCallCheck(this, MdCake);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdCake).apply(this, arguments));
    }

    _createClass(MdCake, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm30 15q2.0333333333333314 0 3.5166666666666657 1.4833333333333343t1.4833333333333343 3.5166666666666657v2.578333333333333q0 1.3283333333333331-0.9766666666666666 2.3049999999999997t-2.3049999999999997 0.9766666666666666-2.2633333333333354-0.9383333333333326l-3.5933333333333337-3.5933333333333337-3.5933333333333337 3.5933333333333337q-0.9383333333333326 0.9383333333333326-2.3049999999999997 0.9383333333333326t-2.3049999999999997-0.9383333333333326l-3.5166666666666657-3.5933333333333337-3.591666666666667 3.5933333333333337q-0.9399999999999995 0.9383333333333326-2.2666666666666675 0.9383333333333326t-2.3049999999999997-0.9766666666666666-0.9783333333333317-2.304999999999996v-2.5783333333333367q0-2.0333333333333314 1.4866666666666664-3.5166666666666657t3.5166666666666675-1.4833333333333343h8.35833333333333v-3.3599999999999994h3.2833333333333314v3.3599999999999994h8.355000000000004z m-2.3433333333333337 11.64q1.7166666666666686 1.7166666666666686 4.063333333333333 1.7166666666666686 1.7166666666666686 0 3.2833333333333314-1.0133333333333319v7.656666666666663q0 0.7033333333333331-0.46999999999999886 1.1716666666666669t-1.1716666666666669 0.46666666666666856h-26.716666666666665q-0.7049999999999992 0-1.173333333333332-0.46666666666666856t-0.4666666666666668-1.1716666666666669v-7.656666666666666q1.4833333333333334 1.0166666666666657 3.2799999999999994 1.0166666666666657 2.3450000000000006 0 4.066666666666666-1.7199999999999989l1.795-1.7966666666666669 1.7966666666666669 1.7966666666666669q1.6399999999999988 1.6400000000000006 4.063333333333334 1.6400000000000006t4.063333333333333-1.6400000000000006l1.8000000000000007-1.7966666666666669z m-7.656666666666666-16.64q-1.3283333333333331 0-2.3433333333333337-1.0166666666666657t-1.0166666666666657-2.341666666666667q0-0.8600000000000003 0.5500000000000007-1.7166666666666668l2.8099999999999987-4.925000000000001 2.8166666666666664 4.923333333333334q0.5450000000000017 0.8600000000000003 0.5450000000000017 1.7166666666666668 0 1.33-0.9766666666666666 2.3450000000000006t-2.3850000000000016 1.0149999999999988z' })
                )
            );
        }
    }]);

    return MdCake;
}(React.Component);

exports.default = MdCake;
module.exports = exports['default'];