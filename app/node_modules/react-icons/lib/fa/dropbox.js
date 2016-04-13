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

var FaDropbox = function (_React$Component) {
    _inherits(FaDropbox, _React$Component);

    function FaDropbox() {
        _classCallCheck(this, FaDropbox);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaDropbox).apply(this, arguments));
    }

    _createClass(FaDropbox, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm8.971428571428572 15.781428571428572l11.028571428571428 6.808571428571428-7.634285714285715 6.361428571428572-10.937142857142856-7.120000000000001z m22.014285714285716 12.389999999999999v2.4085714285714275l-10.93714285714286 6.539999999999999v0.02285714285714846l-0.022857142857141355-0.022857142857141355-0.02571428571428669 0.022857142857141355v-0.022857142857141355l-10.914285714285715-6.539999999999999v-2.4085714285714346l3.2799999999999994 2.1428571428571423 7.634285714285715-6.342857142857138v-0.04285714285714448l0.022857142857141355 0.022857142857141355 0.022857142857141355-0.022857142857141355v0.04285714285714448l7.657142857142858 6.342857142857142z m-18.62-25.538571428571426l7.6342857142857135 6.362857142857141-11.028571428571428 6.785714285714285-7.542857142857143-6.028571428571427z m18.661428571428573 13.147142857142855l7.54428571428571 6.04857142857143-10.914285714285715 7.121428571428574-7.657142857142855-6.361428571428576z m-3.37142857142857-13.147142857142857l10.91571428571428 7.121428571428572-7.5428571428571445 6.0285714285714285-11.028571428571425-6.787142857142857z' })
                )
            );
        }
    }]);

    return FaDropbox;
}(React.Component);

exports.default = FaDropbox;
module.exports = exports['default'];