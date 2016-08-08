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

var MdStopScreenShare = function (_React$Component) {
    _inherits(MdStopScreenShare, _React$Component);

    function MdStopScreenShare() {
        _classCallCheck(this, MdStopScreenShare);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdStopScreenShare).apply(this, arguments));
    }

    _createClass(MdStopScreenShare, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm11.64 25q2.2666666666666675-3.125 6.093333333333334-4.063333333333333l-2.6566666666666663-2.6566666666666663q-2.5 2.421666666666667-3.4383333333333344 6.716666666666669z m-7.656666666666666-22.11l32.891666666666666 32.89000000000001-2.1099999999999994 2.1099999999999994-4.533333333333335-4.533333333333331h-30.231666666666666v-3.3566666666666762h6.638333333333334q-1.4066666666666663 0-2.3433333333333337-0.9783333333333317t-0.9383333333333335-2.3049999999999997v-16.71666666666667q0-1.4866666666666664 1.0933333333333337-2.423333333333334l-2.578333333333333-2.576666666666666z m32.65833333333333 23.826666666666668q0 1.9549999999999983-1.7166666666666686 2.8916666666666657l-9.219999999999999-9.216666666666669 2.6566666666666663-2.5-6.716666666666669-6.173333333333334v3.5166666666666675q-0.158333333333335 0.07833333333333314-0.43333333333333357 0.07833333333333314t-0.42833333333333456 0.07833333333333314l-8.749999999999995-8.674999999999997h21.326666666666668q1.3283333333333331-8.881784197001252e-16 2.3049999999999997 0.9399999999999995t0.9750000000000085 2.343333333333332v16.716666666666665z m-1.25 3.2833333333333314h4.608333333333334v3.3599999999999994h-1.3283333333333331z' })
                )
            );
        }
    }]);

    return MdStopScreenShare;
}(React.Component);

exports.default = MdStopScreenShare;
module.exports = exports['default'];