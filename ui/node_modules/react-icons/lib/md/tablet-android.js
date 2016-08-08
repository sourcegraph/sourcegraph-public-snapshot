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

var MdTabletAndroid = function (_React$Component) {
    _inherits(MdTabletAndroid, _React$Component);

    function MdTabletAndroid() {
        _classCallCheck(this, MdTabletAndroid);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdTabletAndroid).apply(this, arguments));
    }

    _createClass(MdTabletAndroid, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm32.11 31.640000000000004v-26.640000000000004h-24.216666666666665v26.64h24.216666666666665z m-8.75 4.9999999999999964v-1.6400000000000006h-6.716666666666669v1.6400000000000006h6.716666666666669z m6.640000000000001-36.64q2.0333333333333314 0 3.5166666666666657 1.4833333333333334t1.4833333333333343 3.5166666666666666v30q0 2.0333333333333314-1.4833333333333343 3.5166666666666657t-3.5166666666666657 1.4833333333333343h-20q-2.033333333333333 0-3.5166666666666666-1.4833333333333343t-1.4833333333333334-3.5166666666666657v-30q0-2.033333333333333 1.4833333333333334-3.5166666666666666t3.5166666666666666-1.4833333333333334h20z' })
                )
            );
        }
    }]);

    return MdTabletAndroid;
}(React.Component);

exports.default = MdTabletAndroid;
module.exports = exports['default'];