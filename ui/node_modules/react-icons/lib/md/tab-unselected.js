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

var MdTabUnselected = function (_React$Component) {
    _inherits(MdTabUnselected, _React$Component);

    function MdTabUnselected() {
        _classCallCheck(this, MdTabUnselected);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdTabUnselected).apply(this, arguments));
    }

    _createClass(MdTabUnselected, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm28.36 35v-3.3599999999999994h3.2833333333333314v3.3599999999999994h-3.2833333333333314z m-6.719999999999999 0v-3.3599999999999994h3.3599999999999994v3.3599999999999994h-3.3599999999999994z m13.36-13.36v-3.2833333333333314h3.3599999999999994v3.2833333333333314h-3.3599999999999994z m0 13.36v-3.3599999999999994h3.3599999999999994q0 1.3283333333333331-1.0166666666666657 2.3433333333333337t-2.3416666666666686 1.0166666666666657z m-26.64-26.64v-3.3599999999999994h3.283333333333335v3.3599999999999994h-3.283333333333333z m1.7763568394002505e-15 26.64v-3.3599999999999994h3.283333333333333v3.3599999999999994h-3.283333333333333z m6.639999999999999-26.64v-3.3599999999999994h3.3599999999999994v3.3599999999999994h-3.3599999999999994z m20 20v-3.3599999999999994h3.3599999999999994v3.3599999999999994h-3.3599999999999994z m0-23.36q1.3283333333333331 0 2.3433333333333337 1.0166666666666666t1.0166666666666657 2.3416666666666677v6.641666666666666h-16.721666666666664v-10h13.361666666666665z m-30 30q-1.3283333333333331 0-2.3433333333333333-1.0166666666666657t-1.0166666666666666-2.3416666666666686h3.361666666666667v3.3583333333333343z m-3.36-6.640000000000001v-3.3599999999999994h3.36v3.3599999999999994h-3.36z m13.36 6.640000000000001v-3.3599999999999994h3.3599999999999994v3.3599999999999994h-3.3599999999999994z m-13.36-26.64q-4.440892098500626e-16-1.3283333333333314 1.0166666666666662-2.343333333333332t2.3400000000000003-1.0166666666666675v3.3616666666666664h-3.3600000000000003z m-4.440892098500626e-16 13.280000000000001v-3.2833333333333314h3.36v3.2833333333333314h-3.36z m0-6.640000000000001v-3.3599999999999994h3.36v3.3599999999999994h-3.36z' })
                )
            );
        }
    }]);

    return MdTabUnselected;
}(React.Component);

exports.default = MdTabUnselected;
module.exports = exports['default'];