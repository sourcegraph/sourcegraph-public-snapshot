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

var FaProductHunt = function (_React$Component) {
    _inherits(FaProductHunt, _React$Component);

    function FaProductHunt() {
        _classCallCheck(this, FaProductHunt);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaProductHunt).apply(this, arguments));
    }

    _createClass(FaProductHunt, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm25.67142857142857 17.00857142857143q0 1.25-0.8828571428571443 2.120000000000001t-2.131428571428568 0.8714285714285701h-5.647142857142857v-6.007142857142856h5.647142857142857q1.25 0 2.1314285714285717 0.8814285714285717t0.8814285714285717 2.1314285714285717z m3.994285714285713 0q0-2.9000000000000004-2.042857142857141-4.957142857142857t-4.965714285714284-2.0514285714285716h-9.668571428571429v20h4.017142857142858v-6.0042857142857144h5.647142857142857q2.8999999999999986 0 4.9571428571428555-2.0428571428571445t2.0514285714285734-4.942857142857143z m10.334285714285716 2.991428571428571q0 4.062857142857144-1.585714285714289 7.767142857142858t-4.261428571428574 6.385714285714286-6.385714285714286 4.261428571428574-7.764285714285705 1.5857142857142819-7.765714285714285-1.585714285714289-6.385714285714285-4.262857142857143-4.261428571428572-6.385714285714286-1.5857142857142892-7.765714285714282 1.5857142857142859-7.767142857142858 4.261428571428572-6.385714285714285 6.385714285714285-4.261428571428572 7.762857142857143-1.5857142857142854 7.768571428571427 1.5857142857142859 6.385714285714286 4.262857142857143 4.261428571428574 6.385714285714286 1.5842857142857127 7.765714285714285z' })
                )
            );
        }
    }]);

    return FaProductHunt;
}(React.Component);

exports.default = FaProductHunt;
module.exports = exports['default'];