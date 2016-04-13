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

var FaImage = function (_React$Component) {
    _inherits(FaImage, _React$Component);

    function FaImage() {
        _classCallCheck(this, FaImage);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaImage).apply(this, arguments));
    }

    _createClass(FaImage, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm13.333333333333332 13.333333333333332q0 1.666666666666666-1.166666666666666 2.833333333333332t-2.833333333333334 1.1666666666666679-2.833333333333333-1.1666666666666679-1.166666666666666-2.833333333333332 1.166666666666667-2.833333333333332 2.833333333333332-1.1666666666666679 2.833333333333332 1.166666666666666 1.1666666666666679 2.833333333333334z m21.333333333333332 8v9.333333333333332h-29.333333333333332v-4l6.666666666666668-6.666666666666664 3.333333333333332 3.333333333333332 10.666666666666668-10.666666666666666z m2-14.666666666666666h-33.33333333333333q-0.2706666666666666 0-0.46933333333333316 0.19733333333333292t-0.19733333333333603 0.46933333333333405v25.333333333333332q0 0.2706666666666635 0.19733333333333336 0.4693333333333314t0.46933333333333316 0.19733333333333292h33.33333333333333q0.2706666666666635 0 0.4693333333333314-0.19733333333333292t0.19733333333333292-0.4693333333333314v-25.333333333333332q0-0.27066666666666706-0.19733333333333292-0.46933333333333316t-0.4693333333333314-0.19733333333333292z m3.3333333333333357 0.666666666666667v25.333333333333332q0 1.3733333333333348-0.978666666666669 2.3546666666666667t-2.3546666666666667 0.978666666666669h-33.33333333333333q-1.3733333333333333 0-2.3546666666666667-0.978666666666669t-0.978666666666669-2.3546666666666667v-25.333333333333332q0-1.373333333333333 0.9786666666666666-2.3546666666666667t2.3546666666666667-0.9786666666666655h33.33333333333333q1.3733333333333348 0 2.3546666666666667 0.9786666666666664t0.978666666666669 2.3546666666666667z' })
                )
            );
        }
    }]);

    return FaImage;
}(React.Component);

exports.default = FaImage;
module.exports = exports['default'];