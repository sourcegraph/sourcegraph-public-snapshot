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

var MdLocalLaundryService = function (_React$Component) {
    _inherits(MdLocalLaundryService, _React$Component);

    function MdLocalLaundryService() {
        _classCallCheck(this, MdLocalLaundryService);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdLocalLaundryService).apply(this, arguments));
    }

    _createClass(MdLocalLaundryService, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm20 33.36q4.140000000000001 0 7.07-2.9299999999999997t2.9299999999999997-7.07-2.9299999999999997-7.07-7.07-2.929999999999998-7.07 2.929999999999998-2.9299999999999997 7.07 2.9299999999999997 7.07 7.07 2.9299999999999997z m-8.36-26.72q-0.7033333333333331 0-1.1716666666666669 0.5083333333333329t-0.4683333333333337 1.209999999999999 0.4666666666666668 1.1716666666666669 1.1733333333333338 0.4666666666666668 1.211666666666666-0.4666666666666668 0.5099999999999998-1.1733333333333338-0.5083333333333329-1.211666666666667-1.209999999999999-0.5099999999999998z m5 0q-0.7033333333333331 0-1.1716666666666669 0.5083333333333329t-0.4683333333333337 1.209999999999999 0.4666666666666668 1.1716666666666669 1.1733333333333338 0.4666666666666668 1.211666666666666-0.4666666666666668 0.5100000000000016-1.1733333333333338-0.5083333333333329-1.211666666666667-1.2100000000000009-0.5099999999999998z m13.36-3.2800000000000002q1.4066666666666663 0 2.383333333333333 0.9383333333333335t0.9766666666666666 2.341666666666667v26.71666666666667q0 1.4083333333333314-0.9766666666666666 2.344999999999999t-2.383333333333333 0.9383333333333326h-20q-1.4066666666666663 0-2.383333333333333-0.9383333333333326t-0.9766666666666666-2.3433333333333337v-26.715q0-1.408333333333334 0.9766666666666666-2.3450000000000006t2.383333333333333-0.94h20z m-14.686666666666666 24.686666666666667l9.374999999999998-9.453333333333333q1.9533333333333331 1.9533333333333331 1.9533333333333331 4.726666666666667t-1.9533333333333331 4.726666666666667-4.688333333333333 1.9533333333333331-4.688333333333333-1.9533333333333331z' })
                )
            );
        }
    }]);

    return MdLocalLaundryService;
}(React.Component);

exports.default = MdLocalLaundryService;
module.exports = exports['default'];