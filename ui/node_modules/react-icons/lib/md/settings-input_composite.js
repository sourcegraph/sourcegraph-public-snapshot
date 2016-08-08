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

var MdSettingsInputComposite = function (_React$Component) {
    _inherits(MdSettingsInputComposite, _React$Component);

    function MdSettingsInputComposite() {
        _classCallCheck(this, MdSettingsInputComposite);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdSettingsInputComposite).apply(this, arguments));
    }

    _createClass(MdSettingsInputComposite, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm28.36 26.64v-3.2833333333333314h10v3.2833333333333314q0 1.6400000000000006-0.9383333333333326 2.8900000000000006t-2.421666666666667 1.7966666666666669v7.033333333333331h-3.3599999999999994v-7.033333333333331q-3.2833333333333314-1.1716666666666669-3.2833333333333314-4.688333333333333z m-6.719999999999999-23.28v6.640000000000001h3.3599999999999994v10h-10v-10h3.3599999999999994v-6.64q0-0.7033333333333331 0.466666666666665-1.211666666666667t1.1733333333333356-0.51 1.1716666666666669 0.5083333333333331 0.466666666666665 1.21z m13.36 6.640000000000001h3.3599999999999994v10h-10v-10h3.2833333333333314v-6.64q0-0.7033333333333331 0.5066666666666677-1.211666666666667t1.211666666666666-0.51 1.1716666666666669 0.5083333333333331 0.46666666666666856 1.21v6.6433333333333335z m-33.36 16.64v-3.2833333333333314h10v3.2833333333333314q0 3.5166666666666657-3.283333333333333 4.688333333333333v7.033333333333331h-3.3566666666666674v-7.033333333333331q-1.4866666666666668-0.5466666666666669-2.4233333333333333-1.7966666666666669t-0.9383333333333332-2.8916666666666657z m13.360000000000001 0v-3.2833333333333314h9.999999999999998v3.2833333333333314q0 1.6400000000000006-0.9383333333333326 2.8900000000000006t-2.421666666666667 1.7966666666666669v7.033333333333331h-3.2833333333333314v-7.033333333333331q-1.4833333333333343-0.5466666666666669-2.42-1.7966666666666669t-0.9366666666666692-2.8900000000000006z m-6.640000000000001-23.28v6.640000000000001h3.283333333333333v10h-10v-10h3.3566666666666656v-6.64q0-0.7033333333333331 0.46999999999999975-1.211666666666667t1.1716666666666669-0.51 1.211666666666667 0.5083333333333331 0.5083333333333329 1.21z' })
                )
            );
        }
    }]);

    return MdSettingsInputComposite;
}(React.Component);

exports.default = MdSettingsInputComposite;
module.exports = exports['default'];