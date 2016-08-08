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

var FaClose = function (_React$Component) {
    _inherits(FaClose, _React$Component);

    function FaClose() {
        _classCallCheck(this, FaClose);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaClose).apply(this, arguments));
    }

    _createClass(FaClose, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm33.25714285714286 29.50857142857143q0 0.8928571428571423-0.6242857142857119 1.5171428571428578l-3.0357142857142847 3.0357142857142847q-0.6257142857142846 0.6257142857142881-1.5171428571428578 0.6257142857142881t-1.5171428571428578-0.6242857142857119l-6.562857142857148-6.562857142857148-6.562857142857144 6.561428571428571q-0.6257142857142863 0.6257142857142881-1.5171428571428578 0.6257142857142881t-1.5171428571428578-0.6242857142857119l-3.0357142857142856-3.0357142857142847q-0.6257142857142854-0.6257142857142846-0.6257142857142854-1.5171428571428578t0.6242857142857146-1.5171428571428578l6.562857142857145-6.564285714285717-6.561428571428571-6.562857142857144q-0.6257142857142854-0.6257142857142863-0.6257142857142854-1.5171428571428578t0.6242857142857146-1.5171428571428578l3.0357142857142847-3.0357142857142847q0.6257142857142863-0.6257142857142863 1.5171428571428578-0.6257142857142863t1.5171428571428578 0.6242857142857137l6.564285714285713 6.562857142857144 6.562857142857144-6.561428571428571q0.6257142857142846-0.6257142857142863 1.5171428571428578-0.6257142857142863t1.5171428571428578 0.6242857142857137l3.0357142857142883 3.0357142857142847q0.6257142857142881 0.6257142857142863 0.6257142857142881 1.5171428571428578t-0.6242857142857119 1.5171428571428578l-6.562857142857151 6.564285714285717 6.561428571428575 6.562857142857144q0.6257142857142881 0.6257142857142846 0.6257142857142881 1.5171428571428578z' })
                )
            );
        }
    }]);

    return FaClose;
}(React.Component);

exports.default = FaClose;
module.exports = exports['default'];