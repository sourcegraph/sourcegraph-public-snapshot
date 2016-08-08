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

var TiInfinity = function (_React$Component) {
    _inherits(TiInfinity, _React$Component);

    function TiInfinity() {
        _classCallCheck(this, TiInfinity);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiInfinity).apply(this, arguments));
    }

    _createClass(TiInfinity, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm27.388333333333335 14.326666666666668c-1.9216666666666669 0-3.7283333333333317 0.75-5.060000000000002 2.076666666666668l-2.328333333333333 2.2333333333333307-2.289999999999999-2.1999999999999993c-1.3583333333333343-1.3616666666666664-3.166666666666666-2.1099999999999994-5.091666666666667-2.1099999999999994s-3.7300000000000004 0.75-5.086666666666667 2.1099999999999994c-1.3616666666666664 1.3599999999999994-2.1116666666666672 3.166666666666668-2.1116666666666672 5.091666666666669 0 1.9200000000000017 0.75 3.7300000000000004 2.1116666666666664 5.088333333333335 1.3566666666666674 1.3599999999999994 3.166666666666666 2.1099999999999994 5.09 2.1099999999999994s3.7333333333333343-0.75 5.061666666666667-2.080000000000002l2.3249999999999993-2.2300000000000004 2.293333333333333 2.1999999999999993c1.3583333333333343 1.3599999999999994 3.166666666666668 2.1099999999999994 5.091666666666665 2.1099999999999994s3.7300000000000004-0.75 5.088333333333331-2.1099999999999994c1.3616666666666646-1.3566666666666656 2.1116666666666646-3.166666666666668 2.1116666666666646-5.091666666666669s-0.75-3.7300000000000004-2.1099999999999994-5.091666666666669c-1.3616666666666681-1.3583333333333343-3.166666666666668-2.1066666666666674-5.09-2.1066666666666674z m-12.626666666666669 9.34c-1.1449999999999996 1.1499999999999986-3.1400000000000006 1.1499999999999986-4.286666666666667 0-0.5733333333333341-0.5716666666666654-0.8883333333333336-1.3333333333333321-0.8883333333333336-2.1400000000000006 0-0.8099999999999987 0.31666666666666643-1.5666666666666664 0.8916666666666675-2.1449999999999996 0.5700000000000003-0.5749999999999993 1.333333333333334-0.8900000000000006 2.1400000000000006-0.8900000000000006s1.5716666666666672 0.31666666666666643 2.1766666666666676 0.9166666666666679l2.1999999999999993 2.116666666666667-2.2333333333333343 2.1416666666666657z m14.766666666666666 0c-1.1433333333333344 1.1499999999999986-3.1083333333333343 1.1766666666666659-4.316666666666666-0.028333333333332433l-2.1999999999999993-2.116666666666667 2.2333333333333343-2.1416666666666657c1.1449999999999996-1.1466666666666683 3.1416666666666657-1.1466666666666683 4.286666666666665-0.0033333333333338544 0.5716666666666654 0.576666666666668 0.8866666666666667 1.3333333333333321 0.8866666666666667 2.1449999999999996s-0.31666666666666643 1.5733333333333341-0.8900000000000006 2.1466666666666683z' })
                )
            );
        }
    }]);

    return TiInfinity;
}(React.Component);

exports.default = TiInfinity;
module.exports = exports['default'];