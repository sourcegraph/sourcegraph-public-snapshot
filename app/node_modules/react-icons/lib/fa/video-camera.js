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

var FaVideoCamera = function (_React$Component) {
    _inherits(FaVideoCamera, _React$Component);

    function FaVideoCamera() {
        _classCallCheck(this, FaVideoCamera);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaVideoCamera).apply(this, arguments));
    }

    _createClass(FaVideoCamera, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm40 7.857142857142858v24.28571428571428q0 0.9371428571428595-0.8714285714285737 1.317142857142855-0.28857142857143003 0.11142857142856855-0.5571428571428569 0.11142857142856855-0.6028571428571396 0-1.0042857142857144-0.42428571428571615l-8.995714285714282-8.995714285714275v3.7057142857142864q0 2.6571428571428584-1.8857142857142861 4.542857142857141t-4.5428571428571445 1.8857142857142861h-15.714285714285715q-2.6571428571428575 0-4.542857142857144-1.8857142857142861t-1.8857142857142835-4.542857142857141v-15.714285714285715q0-2.6571428571428584 1.885714285714286-4.542857142857144t4.542857142857143-1.8857142857142843h15.714285714285717q2.6571428571428584 0 4.5428571428571445 1.8857142857142861t1.8857142857142826 4.542857142857142v3.6828571428571433l8.995714285714282-8.971428571428572q0.3999999999999986-0.42571428571428527 1.0042857142857144-0.42571428571428527 0.2671428571428578 0 0.5571428571428569 0.11142857142857121 0.8714285714285737 0.3799999999999999 0.8714285714285737 1.3171428571428576z' })
                )
            );
        }
    }]);

    return FaVideoCamera;
}(React.Component);

exports.default = FaVideoCamera;
module.exports = exports['default'];