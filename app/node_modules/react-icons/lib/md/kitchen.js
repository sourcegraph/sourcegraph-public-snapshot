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

var MdKitchen = function (_React$Component) {
    _inherits(MdKitchen, _React$Component);

    function MdKitchen() {
        _classCallCheck(this, MdKitchen);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdKitchen).apply(this, arguments));
    }

    _createClass(MdKitchen, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm13.360000000000001 20h3.283333333333333v8.36h-3.283333333333333v-8.36z m0-11.64h3.283333333333333v5h-3.283333333333333v-5z m16.64 6.640000000000001v-8.36h-20v8.36h20z m0 18.36v-15.076666666666664h-20v15.076666666666664h20z m0-30q1.4066666666666663 0 2.383333333333333 0.9383333333333335t0.9766666666666666 2.341666666666667v26.71666666666667q0 1.3299999999999983-1.0166666666666657 2.306666666666665t-2.3416666666666686 0.9750000000000014h-20.001666666666665q-1.3266666666666662 0-2.341666666666667-0.9766666666666666t-1.0166666666666666-2.306666666666665v-26.71166666666667q0-1.408333333333334 0.9783333333333335-2.3450000000000006t2.3833333333333337-0.94h20z' })
                )
            );
        }
    }]);

    return MdKitchen;
}(React.Component);

exports.default = MdKitchen;
module.exports = exports['default'];