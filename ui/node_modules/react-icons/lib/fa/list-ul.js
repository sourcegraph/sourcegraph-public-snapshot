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

var FaListUl = function (_React$Component) {
    _inherits(FaListUl, _React$Component);

    function FaListUl() {
        _classCallCheck(this, FaListUl);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaListUl).apply(this, arguments));
    }

    _createClass(FaListUl, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm8.571428571428571 31.42857142857143q0 1.7857142857142847-1.25 3.0357142857142847t-3.0357142857142856 1.25-3.0357142857142856-1.25-1.25-3.0357142857142847 1.25-3.0357142857142847 3.0357142857142856-1.2500000000000036 3.0357142857142856 1.25 1.25 3.0357142857142883z m0-11.42857142857143q0 1.7857142857142847-1.25 3.0357142857142847t-3.0357142857142856 1.25-3.0357142857142856-1.25-1.25-3.0357142857142847 1.25-3.0357142857142847 3.0357142857142856-1.25 3.0357142857142856 1.25 1.25 3.0357142857142847z m31.42857142857143 9.285714285714285v4.285714285714285q0 0.28999999999999915-0.21142857142856997 0.5028571428571453t-0.5028571428571453 0.21142857142856997h-27.142857142857142q-0.28999999999999915 0-0.5028571428571436-0.21142857142856997t-0.21142857142856997-0.5028571428571453v-4.285714285714285q0-0.28999999999999915 0.21142857142857174-0.5028571428571418t0.5028571428571418-0.21142857142856997h27.142857142857142q0.28999999999999915 0 0.5028571428571453 0.21142857142856997t0.21142857142856997 0.5028571428571418z m-31.42857142857143-20.714285714285715q1.7763568394002505e-15 1.7857142857142883-1.2499999999999982 3.0357142857142883t-3.0357142857142856 1.25-3.0357142857142856-1.25-1.25-3.0357142857142865 1.25-3.0357142857142856 3.0357142857142856-1.25 3.0357142857142856 1.25 1.25 3.0357142857142856z m31.42857142857143 9.285714285714288v4.285714285714285q0 0.28999999999999915-0.21142857142856997 0.5028571428571418t-0.5028571428571453 0.21142857142857352h-27.142857142857142q-0.28999999999999915 0-0.5028571428571436-0.21142857142856997t-0.21142857142856997-0.5028571428571453v-4.285714285714285q0-0.28999999999999915 0.21142857142857174-0.5028571428571418t0.5028571428571418-0.21142857142857352h27.142857142857142q0.28999999999999915 0 0.5028571428571453 0.21142857142856997t0.21142857142856997 0.5028571428571453z m0-11.428571428571429v4.2857142857142865q0 0.28999999999999915-0.21142857142856997 0.5028571428571436t-0.5028571428571453 0.21142857142856997h-27.142857142857142q-0.28999999999999915 0-0.5028571428571436-0.21142857142857174t-0.21142857142856997-0.5028571428571418v-4.285714285714286q0-0.29000000000000004 0.21142857142857174-0.5028571428571427t0.5028571428571418-0.21142857142857263h27.142857142857142q0.28999999999999915 0 0.5028571428571453 0.21142857142857174t0.21142857142856997 0.5028571428571427z' })
                )
            );
        }
    }]);

    return FaListUl;
}(React.Component);

exports.default = FaListUl;
module.exports = exports['default'];