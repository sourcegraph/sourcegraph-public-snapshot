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

var MdRoomService = function (_React$Component) {
    _inherits(MdRoomService, _React$Component);

    function MdRoomService() {
        _classCallCheck(this, MdRoomService);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdRoomService).apply(this, arguments));
    }

    _createClass(MdRoomService, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm23.046666666666667 12.966666666666667q4.921666666666667 1.0166666666666675 8.283333333333331 4.806666666666667t3.6700000000000017 8.866666666666667h-30q0.31333333333333346-5.078333333333333 3.671666666666667-8.866666666666667t8.283333333333331-4.805q-0.31666666666666643-0.7799999999999994-0.31666666666666643-1.3266666666666662 0-1.3283333333333331 1.0166666666666657-2.3049999999999997t2.3450000000000024-0.9766666666666666 2.3416666666666686 0.9766666666666666 1.0166666666666657 2.3066666666666666q0 0.5466666666666669-0.31666666666666643 1.3283333333333331z m-19.686666666666667 15.395000000000001h33.28333333333333v3.283333333333335h-33.285v-3.2833333333333314z' })
                )
            );
        }
    }]);

    return MdRoomService;
}(React.Component);

exports.default = MdRoomService;
module.exports = exports['default'];