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

var FaFortAwesome = function (_React$Component) {
    _inherits(FaFortAwesome, _React$Component);

    function FaFortAwesome() {
        _classCallCheck(this, FaFortAwesome);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaFortAwesome).apply(this, arguments));
    }

    _createClass(FaFortAwesome, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm15.714285714285715 22.5v-5q0-0.35714285714285765-0.35714285714285765-0.35714285714285765h-2.1428571428571423q-0.35714285714285765 0-0.35714285714285765 0.35714285714285765v5q0 0.35714285714285765 0.35714285714285765 0.35714285714285765h2.1428571428571423q0.35714285714285765 0 0.35714285714285765-0.35714285714285765z m11.42857142857143 0v-5q0-0.35714285714285765-0.35714285714285765-0.35714285714285765h-2.1428571428571423q-0.35714285714285765 0-0.35714285714285765 0.35714285714285765v5q0 0.35714285714285765 0.35714285714285765 0.35714285714285765h2.1428571428571423q0.35714285714285765 0 0.35714285714285765-0.35714285714285765z m11.42857142857143 0.7142857142857153v16.785714285714285h-14.285714285714292v-7.142857142857146q0-1.7857142857142847-1.25-3.0357142857142847t-3.0357142857142847-1.2499999999999964-3.0357142857142847 1.25-1.25 3.0357142857142883v7.142857142857146h-14.285714285714286v-16.785714285714292q-2.220446049250313e-16-0.35714285714285765 0.357142857142857-0.35714285714285765h2.1428571428571432q0.35714285714285676 0 0.35714285714285676 0.35714285714285765v2.5h2.8571428571428568v-13.928571428571429q8.881784197001252e-16-0.35714285714285765 0.35714285714285765-0.35714285714285765h2.1428571428571423q0.35714285714285765 0 0.35714285714285765 0.35714285714285765v2.5h2.8571428571428577v-2.5q0-0.35714285714285765 0.35714285714285765-0.35714285714285765h2.1428571428571423q0.35714285714285765 0 0.35714285714285765 0.35714285714285765v2.5h2.8571428571428577v-2.5q0-0.35714285714285765 0.35714285714285765-0.35714285714285765h0.35714285714285765v-8.771428571428572q-0.7142857142857153-0.42857142857142794-0.7142857142857153-1.228571428571428 0-0.5800000000000001 0.4242857142857126-1.0042857142857142t1.0042857142857144-0.4242857142857144 1.0042857142857144 0.42428571428571427 0.42428571428571615 1.0042857142857144q0 0.8028571428571425-0.7142857142857153 1.2285714285714286v0.19999999999999973h6.071428571428569q0.35714285714285765 0 0.35714285714285765 0.3571428571428572v5.000000000000001q0 0.35714285714285765-0.35714285714285765 0.35714285714285765h-6.071428571428569v2.8571428571428577h0.35714285714285765q0.35714285714285765 0 0.35714285714285765 0.35714285714285765v2.4999999999999982h2.8571428571428577v-2.5q0-0.35714285714285765 0.35714285714285765-0.35714285714285765h2.1428571428571423q0.35714285714285765 0 0.35714285714285765 0.35714285714285765v2.5h2.8571428571428577v-2.5q0-0.35714285714285765 0.35714285714285765-0.35714285714285765h2.142857142857146q0.3571428571428541 0 0.3571428571428541 0.35714285714285765v13.928571428571429h2.857142857142854v-2.5q0-0.35714285714285765 0.3571428571428541-0.35714285714285765h2.142857142857146q0.3571428571428541 0 0.3571428571428541 0.35714285714285765z' })
                )
            );
        }
    }]);

    return FaFortAwesome;
}(React.Component);

exports.default = FaFortAwesome;
module.exports = exports['default'];