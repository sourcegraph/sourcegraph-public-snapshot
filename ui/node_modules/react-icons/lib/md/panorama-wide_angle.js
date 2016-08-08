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

var MdPanoramaWideAngle = function (_React$Component) {
    _inherits(MdPanoramaWideAngle, _React$Component);

    function MdPanoramaWideAngle() {
        _classCallCheck(this, MdPanoramaWideAngle);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdPanoramaWideAngle).apply(this, arguments));
    }

    _createClass(MdPanoramaWideAngle, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm20 6.640000000000001q6.016666666666666 0 13.283333333333331 1.25l1.4833333333333343 0.2333333333333325 0.46666666666666856 1.4866666666666664q1.4083333333333314 5.156666666666666 1.4083333333333314 10.39t-1.4066666666666663 10.39l-0.46666666666666856 1.4833333333333343-1.4833333333333343 0.23666666666666458q-7.266666666666666 1.25-13.283333333333331 1.25t-13.283333333333333-1.25l-1.4833333333333334-0.23333333333333428-0.4666666666666668-1.4866666666666681q-1.4116666666666653-5.156666666666663-1.4116666666666653-10.389999999999997t1.4066666666666672-10.39l0.4666666666666668-1.4833333333333343 1.4833333333333334-0.23666666666666636q7.266666666666667-1.25 13.283333333333331-1.25z m0 3.3599999999999994q-5.466666666666667 0-12.188333333333333 1.0933333333333337-1.1716666666666669 4.453333333333333-1.1716666666666669 8.906666666666666t1.1716666666666669 8.906666666666666q6.716666666666667 1.0933333333333337 12.188333333333333 1.0933333333333337t12.188333333333333-1.0933333333333337q1.1716666666666669-4.453333333333333 1.1716666666666669-8.906666666666666t-1.1716666666666669-8.906666666666668q-6.716666666666665-1.093333333333332-12.188333333333333-1.093333333333332z' })
                )
            );
        }
    }]);

    return MdPanoramaWideAngle;
}(React.Component);

exports.default = MdPanoramaWideAngle;
module.exports = exports['default'];