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

var FaLinkedinSquare = function (_React$Component) {
    _inherits(FaLinkedinSquare, _React$Component);

    function FaLinkedinSquare() {
        _classCallCheck(this, FaLinkedinSquare);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaLinkedinSquare).apply(this, arguments));
    }

    _createClass(FaLinkedinSquare, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm8.147142857142857 31.562857142857144h5.157142857142858v-15.491428571428571h-5.157142857142858v15.491428571428571z m5.491428571428573-20.26857142857143q-0.02285714285714313-1.1600000000000001-0.8028571428571425-1.92t-2.0757142857142856-0.757142857142858-2.1099999999999994 0.757142857142858-0.8142857142857141 1.92q0 1.138571428571428 0.7928571428571436 1.9085714285714293t2.064285714285715 0.7714285714285722h0.021428571428568688q1.3185714285714294 0 2.121428571428572-0.7714285714285722t0.8028571428571425-1.9085714285714293z m13.057142857142859 20.26857142857143h5.157142857142858v-8.885714285714286q0-3.435714285714287-1.62857142857143-5.199999999999999t-4.310000000000002-1.7628571428571433q-3.0357142857142847 0-4.665714285714284 2.611428571428572h0.04285714285714448v-2.2542857142857144h-5.152857142857144q0.0671428571428585 1.4714285714285715 0 15.491428571428571h5.157142857142858v-8.66q0-0.8485714285714288 0.15428571428571303-1.25 0.33428571428571274-0.7814285714285703 1.0042857142857144-1.3285714285714292t1.6514285714285712-0.5471428571428589q2.59 0 2.59 3.5042857142857144v8.28142857142857z m10.447142857142858-22.277142857142856v21.428571428571427q0 2.6571428571428584-1.8857142857142861 4.5428571428571445t-4.5428571428571445 1.8857142857142861h-21.42857142857143q-2.6571428571428575 0-4.542857142857144-1.8857142857142861t-1.885714285714284-4.5428571428571445v-21.42857142857143q0-2.6571428571428575 1.8857142857142857-4.542857142857144t4.542857142857144-1.885714285714284h21.42857142857143q2.6571428571428584 0 4.5428571428571445 1.8857142857142857t1.8857142857142861 4.542857142857144z' })
                )
            );
        }
    }]);

    return FaLinkedinSquare;
}(React.Component);

exports.default = FaLinkedinSquare;
module.exports = exports['default'];