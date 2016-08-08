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

var MdPanoramaVertical = function (_React$Component) {
    _inherits(MdPanoramaVertical, _React$Component);

    function MdPanoramaVertical() {
        _classCallCheck(this, MdPanoramaVertical);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdPanoramaVertical).apply(this, arguments));
    }

    _createClass(MdPanoramaVertical, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm10.938333333333333 33.36h18.125q-1.875-6.483333333333334-1.875-13.36t1.875-13.360000000000001h-18.125q1.875 6.406666666666668 1.875 13.360000000000001 0 6.875-1.875 13.36z m22.266666666666666 1.875q0.15500000000000114 0.31666666666667 0.15500000000000114 0.46999999999999886 0 0.9383333333333326-1.0933333333333337 0.9383333333333326h-24.53333333333333q-1.0916666666666677 0-1.0916666666666677-0.9383333333333326 0-0.1566666666666663 0.1566666666666663-0.46666666666666856 2.7350000000000003-7.5049999999999955 2.7350000000000003-15.23833333333333t-2.7333333333333325-15.233333333333334q-0.15833333333333321-0.31666666666666554-0.15833333333333321-0.46999999999999886 0-0.9383333333333335 1.0933333333333337-0.9383333333333335h24.53333333333334q1.0933333333333337 0 1.0933333333333337 0.9383333333333335 0 0.1566666666666663-0.1566666666666663 0.4666666666666668-2.7333333333333343 7.5-2.7333333333333343 15.236666666666666t2.7333333333333343 15.233333333333334z' })
                )
            );
        }
    }]);

    return MdPanoramaVertical;
}(React.Component);

exports.default = MdPanoramaVertical;
module.exports = exports['default'];