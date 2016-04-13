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

var FaStickyNoteO = function (_React$Component) {
    _inherits(FaStickyNoteO, _React$Component);

    function FaStickyNoteO() {
        _classCallCheck(this, FaStickyNoteO);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaStickyNoteO).apply(this, arguments));
    }

    _createClass(FaStickyNoteO, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm34.10714285714286 28.571428571428573h-5.535714285714288v5.535714285714288q0.6471428571428568-0.2228571428571442 0.9142857142857146-0.4914285714285711l4.131428571428572-4.12857142857143q0.2671428571428578-0.2671428571428578 0.4914285714285711-0.9142857142857146z m-6.2500000000000036-2.8571428571428577h6.428571428571427v-20h-28.57142857142857v28.57142857142857h20v-6.428571428571431q0-0.8928571428571423 0.6257142857142846-1.5171428571428578t1.5171428571428578-0.625714285714281z m9.285714285714288-20.714285714285715v22.857142857142858q0 0.8928571428571423-0.4471428571428575 1.9642857142857153t-1.0714285714285694 1.6971428571428575l-4.107142857142858 4.107142857142858q-0.6257142857142846 0.6257142857142881-1.6971428571428575 1.0714285714285694t-1.9628571428571462 0.4457142857142884h-22.857142857142858q-0.8942857142857141 0-1.5185714285714287-0.6242857142857119t-0.6242857142857141-1.518571428571434v-30q0-0.8914285714285715 0.6242857142857141-1.5142857142857142t1.5185714285714287-0.6285714285714286h30q0.8914285714285697 0 1.5142857142857125 0.6285714285714286t0.6285714285714334 1.5142857142857142z' })
                )
            );
        }
    }]);

    return FaStickyNoteO;
}(React.Component);

exports.default = FaStickyNoteO;
module.exports = exports['default'];