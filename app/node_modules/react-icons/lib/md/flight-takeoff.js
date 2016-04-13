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

var MdFlightTakeoff = function (_React$Component) {
    _inherits(MdFlightTakeoff, _React$Component);

    function MdFlightTakeoff() {
        _classCallCheck(this, MdFlightTakeoff);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdFlightTakeoff).apply(this, arguments));
    }

    _createClass(MdFlightTakeoff, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm36.79666666666667 16.093333333333334q0.23333333333333428 1.0166666666666657-0.27333333333333343 1.875t-1.5233333333333334 1.173333333333332q-9.688333333333333 2.578333333333333-16.093333333333334 4.296666666666667l-8.828333333333333 2.3433333333333337-2.6566666666666663 0.783333333333335-4.375-7.5 2.421666666666667-0.6266666666666652 3.283333333333333 2.5 8.283333333333333-2.1883333333333326-6.876666666666669-11.955 3.203333333333333-0.8600000000000003 11.483333333333334 10.705000000000002 8.908333333333331-2.3433333333333337q1.0166666666666657-0.3133333333333326 1.9166666666666643 0.2333333333333325t1.1333333333333329 1.5666666666666682z m-32.656666666666666 15.54666666666667h31.71666666666667v3.359999999999996h-31.715000000000003v-3.3599999999999994z' })
                )
            );
        }
    }]);

    return MdFlightTakeoff;
}(React.Component);

exports.default = MdFlightTakeoff;
module.exports = exports['default'];