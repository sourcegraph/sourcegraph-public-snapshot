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

var FaShareAltSquare = function (_React$Component) {
    _inherits(FaShareAltSquare, _React$Component);

    function FaShareAltSquare() {
        _classCallCheck(this, FaShareAltSquare);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaShareAltSquare).apply(this, arguments));
    }

    _createClass(FaShareAltSquare, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm31.42857142857143 26.674285714285716q0-1.9642857142857153-1.3957142857142841-3.37142857142857t-3.3599999999999994-1.404285714285713q-1.8742857142857154 0-3.2357142857142875 1.2928571428571445l-5.380000000000003-2.677142857142865q0.04285714285714448-0.35714285714285765 0.04285714285714448-0.5142857142857125t-0.04285714285714448-0.5114285714285707l5.37857142857143-2.678571428571427q1.361428571428572 1.2942857142857136 3.2371428571428567 1.2942857142857136 1.9657142857142844 0 3.361428571428572-1.4057142857142857t1.394285714285715-3.370000000000001-1.3957142857142841-3.3599999999999994-3.361428571428572-1.3971428571428586-3.3685714285714283 1.3971428571428568-1.4057142857142857 3.3599999999999994q0 0.15714285714285658 0.04285714285714448 0.5142857142857142l-5.37857142857143 2.677142857142858q-1.3857142857142861-1.2714285714285722-3.2371428571428567-1.2714285714285722-1.9642857142857153 0-3.3599999999999994 1.394285714285715t-1.395714285714286 3.3599999999999994 1.395714285714286 3.3599999999999994 3.3599999999999994 1.3957142857142841q1.8528571428571432 0 3.2371428571428567-1.2714285714285722l5.379999999999999 2.678571428571427q-0.04571428571428626 0.35714285714285765-0.04571428571428626 0.5142857142857125 0 1.9628571428571426 1.4057142857142857 3.3571428571428577t3.37142857142857 1.3971428571428568 3.3571428571428577-1.3957142857142841 1.3971428571428568-3.3599999999999994z m5.714285714285715-17.38857142857143v21.42857142857143q0 2.6571428571428584-1.8857142857142861 4.5428571428571445t-4.5428571428571445 1.8857142857142861h-21.42857142857143q-2.6571428571428575 0-4.542857142857144-1.8857142857142861t-1.885714285714284-4.5428571428571445v-21.42857142857143q0-2.6571428571428575 1.8857142857142857-4.542857142857144t4.542857142857144-1.885714285714284h21.42857142857143q2.6571428571428584 0 4.5428571428571445 1.8857142857142857t1.8857142857142861 4.542857142857144z' })
                )
            );
        }
    }]);

    return FaShareAltSquare;
}(React.Component);

exports.default = FaShareAltSquare;
module.exports = exports['default'];