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

var FaPaperPlane = function (_React$Component) {
    _inherits(FaPaperPlane, _React$Component);

    function FaPaperPlane() {
        _classCallCheck(this, FaPaperPlane);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaPaperPlane).apply(this, arguments));
    }

    _createClass(FaPaperPlane, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm39.37571428571429 0.2457142857142857q0.7371428571428567 0.5357142857142857 0.6028571428571396 1.4285714285714286l-5.714285714285715 34.285714285714285q-0.11142857142856855 0.6471428571428604-0.7142857142857153 1.0042857142857144-0.3142857142857096 0.1785714285714306-0.692857142857136 0.1785714285714306-0.24285714285714022 0-0.5342857142857156-0.11142857142856855l-10.111428571428572-4.128571428571426-5.399999999999999 6.582857142857144q-0.4042857142857166 0.5142857142857054-1.0971428571428596 0.5142857142857054-0.28857142857142826 0-0.4900000000000002-0.09142857142857252-0.4242857142857144-0.15714285714285836-0.6814285714285706-0.5242857142857176t-0.257142857142858-0.8128571428571405v-7.791428571428572l19.285714285714285-23.637142857142855-23.86142857142857 20.642857142857142-8.817142857142859-3.614285714285714q-0.8257142857142852-0.31428571428571317-0.8928571428571423-1.2285714285714278-0.04285714285714286-0.8928571428571423 0.7142857142857143-1.3171428571428585l37.142857142857146-21.42857142857143q0.33428571428571274-0.19714285714285396 0.7142857142857082-0.19714285714285396 0.4471428571428575 0 0.8028571428571425 0.24285714285714288z' })
                )
            );
        }
    }]);

    return FaPaperPlane;
}(React.Component);

exports.default = FaPaperPlane;
module.exports = exports['default'];