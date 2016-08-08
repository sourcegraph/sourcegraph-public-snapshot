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

var FaMicrophone = function (_React$Component) {
    _inherits(FaMicrophone, _React$Component);

    function FaMicrophone() {
        _classCallCheck(this, FaMicrophone);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaMicrophone).apply(this, arguments));
    }

    _createClass(FaMicrophone, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm32.85714285714286 15.714285714285715v2.8571428571428577q0 4.9328571428571415-3.2928571428571445 8.582857142857144t-8.135714285714286 4.185714285714283v2.945714285714285h5.714285714285715q0.5799999999999983 0 1.0042857142857144 0.42571428571428527t0.4242857142857126 1.0028571428571453-0.4242857142857126 1.0057142857142836-1.004285714285718 0.42285714285714704h-14.285714285714285q-0.5800000000000001 0-1.0042857142857144-0.42285714285713993t-0.4242857142857144-1.0057142857142907 0.4242857142857144-1.0028571428571453 1.0042857142857144-0.42571428571428527h5.714285714285715v-2.9428571428571395q-4.842857142857143-0.5357142857142847-8.135714285714286-4.185714285714287t-3.2928571428571436-8.585714285714285v-2.8571428571428577q0-0.5800000000000001 0.4242857142857144-1.0042857142857144t1.0042857142857136-0.4242857142857144 1.0042857142857144 0.4242857142857144 0.4242857142857144 1.0042857142857144v2.8571428571428577q0 4.12857142857143 2.935714285714287 7.064285714285713t7.064285714285713 2.935714285714287 7.064285714285717-2.935714285714287 2.9357142857142833-7.064285714285713v-2.8571428571428577q0-0.5800000000000001 0.4242857142857126-1.0042857142857144t1.004285714285718-0.4242857142857144 1.0042857142857144 0.4242857142857144 0.42428571428571615 1.0042857142857144z m-5.714285714285715-8.571428571428571v11.428571428571429q0 2.9471428571428575-2.1000000000000014 5.042857142857141t-5.0428571428571445 2.1000000000000014-5.042857142857143-2.1000000000000014-2.0999999999999996-5.042857142857141v-11.42857142857143q0-2.9471428571428566 2.0999999999999996-5.042857142857142t5.042857142857143-2.1000000000000005 5.0428571428571445 2.1 2.099999999999998 5.042857142857143z' })
                )
            );
        }
    }]);

    return FaMicrophone;
}(React.Component);

exports.default = FaMicrophone;
module.exports = exports['default'];