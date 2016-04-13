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

var FaMousePointer = function (_React$Component) {
    _inherits(FaMousePointer, _React$Component);

    function FaMousePointer() {
        _classCallCheck(this, FaMousePointer);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaMousePointer).apply(this, arguments));
    }

    _createClass(FaMousePointer, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm32.432857142857145 23.281428571428574q0.691428571428574 0.6714285714285708 0.3142857142857167 1.5399999999999991-0.3828571428571479 0.8928571428571423-1.3185714285714312 0.8928571428571423h-8.528571428571428l4.488571428571429 10.625714285714288q0.2228571428571442 0.5571428571428569 0 1.0942857142857179t-0.7571428571428562 0.7814285714285703l-3.9528571428571446 1.6757142857142853q-0.5571428571428569 0.2228571428571442-1.0942857142857143 0t-0.7828571428571429-0.7571428571428598l-4.262857142857143-10.091428571428573-6.964285714285715 6.962857142857146q-0.4228571428571435 0.42285714285713993-1.0028571428571436 0.42285714285713993-0.2671428571428578 0-0.5357142857142865-0.10999999999999943-0.8928571428571415-0.38000000000000256-0.8928571428571415-1.3185714285714312v-33.57142857142857q0-0.9342857142857164 0.8928571428571432-1.3142857142857163 0.2671428571428578-0.11428571428571432 0.5357142857142847-0.11428571428571432 0.6028571428571432 0 1.0042857142857144 0.4257142857142857z' })
                )
            );
        }
    }]);

    return FaMousePointer;
}(React.Component);

exports.default = FaMousePointer;
module.exports = exports['default'];