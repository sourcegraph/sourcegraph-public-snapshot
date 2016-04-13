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

var MdPanoramaHorizontal = function (_React$Component) {
    _inherits(MdPanoramaHorizontal, _React$Component);

    function MdPanoramaHorizontal() {
        _classCallCheck(this, MdPanoramaHorizontal);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdPanoramaHorizontal).apply(this, arguments));
    }

    _createClass(MdPanoramaHorizontal, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm35.70333333333333 6.640000000000001q0.9383333333333326 0 0.9383333333333326 1.0933333333333337v24.53333333333333q0 1.0916666666666686-0.9383333333333326 1.0916666666666686-0.1566666666666663 0-0.46666666666666856-0.1566666666666663-7.5-2.7333333333333343-15.236666666666665-2.7333333333333343t-15.233333333333334 2.7333333333333343q-0.31666666666666643 0.15833333333333144-0.46999999999999975 0.15833333333333144-0.9383333333333335 0-0.9383333333333335-1.0933333333333337v-24.53333333333333q0-1.0933333333333346 0.9383333333333335-1.0933333333333346 0.1566666666666663 0 0.4666666666666668 0.1566666666666663 7.499999999999999 2.7333333333333334 15.236666666666668 2.7333333333333334t15.233333333333334-2.7333333333333334q0.31666666666667-0.1566666666666663 0.46999999999999886-0.1566666666666663z m-2.3433333333333337 4.300000000000001q-6.093333333333334 1.8716666666666661-13.36 1.8716666666666661-6.875 0-13.360000000000001-1.875v18.125q6.406666666666668-1.8733333333333348 13.360000000000001-1.8733333333333348 6.875 0 13.36 1.875v-18.125z' })
                )
            );
        }
    }]);

    return MdPanoramaHorizontal;
}(React.Component);

exports.default = MdPanoramaHorizontal;
module.exports = exports['default'];