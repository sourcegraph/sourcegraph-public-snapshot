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

var MdCardMembership = function (_React$Component) {
    _inherits(MdCardMembership, _React$Component);

    function MdCardMembership() {
        _classCallCheck(this, MdCardMembership);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdCardMembership).apply(this, arguments));
    }

    _createClass(MdCardMembership, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm33.36 16.64v-10h-26.716666666666665v10h26.716666666666665z m0 8.36v-3.3599999999999994h-26.716666666666665v3.3599999999999994h26.716666666666665z m0-21.64q1.4066666666666663 0 2.3433333333333337 0.9383333333333335t0.9383333333333326 2.341666666666667v18.36q0 1.408333333333335-0.9383333333333326 2.383333333333333t-2.3433333333333337 0.9783333333333317h-6.716666666666669v8.283333333333331l-6.643333333333331-3.2849999999999966-6.638333333333334 3.2833333333333314v-8.283333333333331h-6.716666666666668q-1.4083333333333332 0-2.3449999999999998-0.9766666666666666t-0.9383333333333321-2.383333333333333v-18.356666666666666q0-1.406666666666668 0.938333333333333-2.3433333333333337t2.343333333333333-0.94h26.71666666666667z' })
                )
            );
        }
    }]);

    return MdCardMembership;
}(React.Component);

exports.default = MdCardMembership;
module.exports = exports['default'];