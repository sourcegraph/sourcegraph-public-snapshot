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

var FaTumblr = function (_React$Component) {
    _inherits(FaTumblr, _React$Component);

    function FaTumblr() {
        _classCallCheck(this, FaTumblr);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaTumblr).apply(this, arguments));
    }

    _createClass(FaTumblr, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm28.92857142857143 29.665714285714284l1.7857142857142847 5.291428571428572q-0.5142857142857125 0.7800000000000011-2.4771428571428586 1.471428571428575t-3.951428571428572 0.7142857142857153q-2.321428571428573 0.04285714285714448-4.252857142857142-0.5799999999999983t-3.1814285714285724-1.6514285714285748-2.120000000000001-2.365714285714283-1.2385714285714293-2.6785714285714306-0.36857142857142833-2.6342857142857135v-12.142857142857142h-3.7499999999999964v-4.800000000000001q1.6071428571428559-0.5757142857142892 2.879999999999999-1.547142857142859t2.031428571428572-2.0114285714285725 1.2942857142857136-2.2771428571428576 0.757142857142858-2.21 0.3371428571428581-1.9728571428571424q0.024285714285714022-0.11142857142857132 0.10000000000000142-0.1899999999999999t0.16857142857142549-0.08142857142857146h5.4471428571428575v9.464285714285715h7.434285714285714v5.624285714285715h-7.4571428571428555v11.562857142857144q0 0.6714285714285708 0.1471428571428568 1.25t0.5028571428571418 1.1714285714285708 1.1028571428571432 0.928571428571427 1.8185714285714276 0.3114285714285714q1.7428571428571438-0.04285714285714448 2.991428571428571-0.6471428571428568z' })
                )
            );
        }
    }]);

    return FaTumblr;
}(React.Component);

exports.default = FaTumblr;
module.exports = exports['default'];