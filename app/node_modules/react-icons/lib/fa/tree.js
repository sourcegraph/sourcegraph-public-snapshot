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

var FaTree = function (_React$Component) {
    _inherits(FaTree, _React$Component);

    function FaTree() {
        _classCallCheck(this, FaTree);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaTree).apply(this, arguments));
    }

    _createClass(FaTree, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm36.42857142857143 32.85714285714286q0 0.5799999999999983-0.42428571428571615 1.0042857142857144t-1.0042857142857144 0.42428571428570905h-10.314285714285717q0.024285714285714022 0.38000000000000256 0.13571428571428612 1.952857142857141t0.1114285714285721 2.421428571428571q0 0.5571428571428569-0.3999999999999986 0.9485714285714266t-0.96142857142857 0.3914285714285768h-7.142857142857142q-0.5571428571428569 0-0.9600000000000009-0.3914285714285697t-0.40000000000000036-0.9485714285714266q0-0.8485714285714252 0.10999999999999943-2.421428571428571t0.13571428571428612-1.9528571428571482h-10.314285714285715q-0.5800000000000001 0-1.0042857142857144-0.42428571428571615t-0.42428571428571393-1.0042857142857073 0.4242857142857144-1.0042857142857144l8.971428571428572-8.995714285714289h-5.109999999999999q-0.5800000000000001 0-1.0042857142857144-0.4242857142857126t-0.4242857142857144-1.0042857142857144 0.4242857142857144-1.0042857142857144l8.971428571428572-8.995714285714287h-4.395714285714286q-0.5800000000000001 0-1.0042857142857144-0.4242857142857144t-0.4242857142857144-1.0042857142857144 0.4242857142857144-1.0042857142857144l8.571428571428571-8.571428571428571q0.4242857142857126-0.4242857142857144 1.0042857142857144-0.4242857142857144t1.0042857142857144 0.42428571428571427l8.57142857142857 8.571428571428571q0.42428571428571615 0.4242857142857144 0.42428571428571615 1.0042857142857144t-0.4242857142857126 1.0042857142857144-1.0042857142857144 0.4242857142857144h-4.397142857142857l8.971428571428575 8.995714285714287q0.42571428571428527 0.4242857142857126 0.42571428571428527 1.0042857142857144t-0.42428571428571615 1.0042857142857144-1.0042857142857144 0.4242857142857126h-5.111428571428572l8.971428571428572 8.995714285714286q0.42571428571428527 0.4242857142857126 0.42571428571428527 1.004285714285711z' })
                )
            );
        }
    }]);

    return FaTree;
}(React.Component);

exports.default = FaTree;
module.exports = exports['default'];