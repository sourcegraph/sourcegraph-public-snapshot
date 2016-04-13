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

var MdVolumeUp = function (_React$Component) {
    _inherits(MdVolumeUp, _React$Component);

    function MdVolumeUp() {
        _classCallCheck(this, MdVolumeUp);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdVolumeUp).apply(this, arguments));
    }

    _createClass(MdVolumeUp, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm23.36 5.390000000000001q5.078333333333333 1.0933333333333337 8.36 5.195t3.2833333333333314 9.416666666666668-3.2833333333333314 9.411666666666665-8.36 5.195v-3.4416666666666664q3.671666666666667-1.091666666666665 5.976666666666667-4.138333333333332t2.3049999999999997-7.033333333333331-2.3049999999999997-7.030000000000001-5.976666666666667-4.141666666666666v-3.43666666666667z m4.140000000000001 14.61q0 4.688333333333333-4.140000000000001 6.716666666666669v-13.433333333333335q4.140000000000001 2.033333333333333 4.140000000000001 6.716666666666667z m-22.5-5h6.640000000000001l8.36-8.36v26.71666666666667l-8.36-8.35666666666667h-6.640000000000001v-10z' })
                )
            );
        }
    }]);

    return MdVolumeUp;
}(React.Component);

exports.default = MdVolumeUp;
module.exports = exports['default'];