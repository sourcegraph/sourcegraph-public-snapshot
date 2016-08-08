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

var FaStackOverflow = function (_React$Component) {
    _inherits(FaStackOverflow, _React$Component);

    function FaStackOverflow() {
        _classCallCheck(this, FaStackOverflow);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaStackOverflow).apply(this, arguments));
    }

    _createClass(FaStackOverflow, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm31.62857142857143 36.42857142857143h-24.952857142857145v-10.714285714285715h-3.5714285714285707v14.285714285714285h32.1v-14.285714285714285h-3.571428571428573v10.714285714285715z m-21.024285714285718-11.695714285714288l0.7371428571428584-3.5042857142857144 17.47714285714286 3.6828571428571415-0.7371428571428567 3.482857142857142z m2.3000000000000025-8.348571428571429l1.4942857142857147-3.26 16.18285714285714 7.567142857142857-1.4957142857142856 3.2371428571428567z m4.485714285714286-7.948571428571428l2.277142857142856-2.7457142857142856 13.704285714285717 11.452857142857143-2.277142857142856 2.7457142857142856z m8.86-8.435714285714285l10.64714285714286 14.30857142857143-2.857142857142854 2.1428571428571423-10.647142857142864-14.30857142857143z m-16.02857142857143 32.83428571428572v-3.548571428571435h17.857142857142858v3.548571428571428h-17.857142857142858z' })
                )
            );
        }
    }]);

    return FaStackOverflow;
}(React.Component);

exports.default = FaStackOverflow;
module.exports = exports['default'];