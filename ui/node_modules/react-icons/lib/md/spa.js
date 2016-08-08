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

var MdSpa = function (_React$Component) {
    _inherits(MdSpa, _React$Component);

    function MdSpa() {
        _classCallCheck(this, MdSpa);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdSpa).apply(this, arguments));
    }

    _createClass(MdSpa, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm25.783333333333335 16.016666666666666q-3.283333333333335 1.7933333333333366-5.783333333333335 4.449999999999999-2.5-2.655000000000001-5.783333333333335-4.449999999999999 0.6266666666666669-7.423333333333332 5.861666666666668-12.658333333333333 5.233333333333334 5.2333333333333325 5.703333333333333 12.656666666666668z m-22.423333333333336 0.6233333333333348q5.313333333333333 0 9.688333333333333 2.578333333333333t6.951666666666668 6.565000000000001q2.5799999999999983-3.986666666666668 6.954999999999998-6.566666666666666t9.688333333333333-2.576666666666668q0 6.563333333333333-3.711666666666666 11.796666666666667t-9.649999999999999 7.343333333333334q-1.3283333333333331 0.46666666666666856-3.2833333333333314 0.8599999999999994-1.6383333333333319-0.23333333333333428-3.280000000000001-0.8599999999999994-5.938333333333333-2.1099999999999994-9.65-7.343333333333334t-3.7066666666666666-11.796666666666667z' })
                )
            );
        }
    }]);

    return MdSpa;
}(React.Component);

exports.default = MdSpa;
module.exports = exports['default'];