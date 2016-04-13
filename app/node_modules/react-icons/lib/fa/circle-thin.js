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

var FaCircleThin = function (_React$Component) {
    _inherits(FaCircleThin, _React$Component);

    function FaCircleThin() {
        _classCallCheck(this, FaCircleThin);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaCircleThin).apply(this, arguments));
    }

    _createClass(FaCircleThin, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm20 5.714285714285714q-2.8999999999999986 0-5.547142857142857 1.1385714285714288t-4.5528571428571425 3.047142857142857-3.047142857142857 4.5528571428571425-1.1385714285714288 5.547142857142857 1.1385714285714288 5.547142857142859 3.047142857142857 4.5528571428571425 4.5528571428571425 3.047142857142859 5.547142857142857 1.1385714285714243 5.547142857142859-1.1385714285714315 4.5528571428571425-3.047142857142859 3.047142857142859-4.5528571428571425 1.1385714285714243-5.547142857142852-1.1385714285714315-5.547142857142857-3.047142857142852-4.5528571428571425-4.5528571428571425-3.047142857142857-5.547142857142859-1.1385714285714288z m17.14285714285714 14.285714285714285q0 4.665714285714287-2.299999999999997 8.604285714285716t-6.237142857142857 6.238571428571426-8.605714285714285 2.3000000000000043-8.6-2.3000000000000043-6.242857142857143-6.238571428571426-2.295714285714286-8.604285714285716 2.3000000000000003-8.604285714285714 6.234285714285714-6.238571428571428 8.604285714285714-2.3000000000000003 8.605714285714285 2.3000000000000003 6.238571428571426 6.238571428571428 2.298571428571435 8.604285714285714z' })
                )
            );
        }
    }]);

    return FaCircleThin;
}(React.Component);

exports.default = FaCircleThin;
module.exports = exports['default'];