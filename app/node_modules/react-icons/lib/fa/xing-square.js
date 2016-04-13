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

var FaXingSquare = function (_React$Component) {
    _inherits(FaXingSquare, _React$Component);

    function FaXingSquare() {
        _classCallCheck(this, FaXingSquare);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaXingSquare).apply(this, arguments));
    }

    _createClass(FaXingSquare, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm18.147142857142857 17.075714285714284q0-0.022857142857141355-2.814285714285715-4.957142857142857-0.4671428571428571-0.757142857142858-1.1571428571428566-0.757142857142858h-4.109999999999999q-0.40000000000000036 0-0.5800000000000001 0.24571428571428555-0.1542857142857148 0.2671428571428578 0.024285714285714022 0.6471428571428568l2.790000000000001 4.821428571428573v0.022857142857141355l-4.375714285714286 7.7214285714285715q-0.20000000000000018 0.31428571428571317 0 0.6257142857142846 0.17857142857142883 0.28999999999999915 0.5357142857142865 0.28999999999999915h4.12857142857143q0.6928571428571431 0 1.1171428571428574-0.8028571428571425z m13.928571428571427-11.092857142857143q-0.1571428571428548-0.268571428571426-0.5357142857142847-0.268571428571426h-4.174285714285716q-0.6714285714285708 0-1.0942857142857143 0.7828571428571429l-9.174285714285713 16.271428571428572q0.022857142857141355 0.04571428571428626 5.848571428571429 10.73857142857143 0.4471428571428575 0.7814285714285703 1.1600000000000001 0.7814285714285703h4.107142857142858q0.3999999999999986 0 0.5571428571428569-0.2671428571428578 0.17857142857142705-0.28999999999999915-0.022857142857141355-0.6257142857142881l-5.8028571428571425-10.625714285714288v-0.024285714285714022l9.12857142857143-16.13857142857143q0.17999999999999972-0.35714285714285676 0-0.6257142857142854z m5.067142857142862 3.302857142857146v21.42857142857143q0 2.6571428571428584-1.8857142857142861 4.5428571428571445t-4.5428571428571445 1.8857142857142861h-21.42857142857143q-2.6571428571428575 0-4.542857142857144-1.8857142857142861t-1.885714285714284-4.5428571428571445v-21.42857142857143q0-2.6571428571428575 1.8857142857142857-4.542857142857144t4.542857142857144-1.885714285714284h21.42857142857143q2.6571428571428584 0 4.5428571428571445 1.8857142857142857t1.8857142857142861 4.542857142857144z' })
                )
            );
        }
    }]);

    return FaXingSquare;
}(React.Component);

exports.default = FaXingSquare;
module.exports = exports['default'];