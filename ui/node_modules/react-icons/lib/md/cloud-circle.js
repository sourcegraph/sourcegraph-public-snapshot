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

var MdCloudCircle = function (_React$Component) {
    _inherits(MdCloudCircle, _React$Component);

    function MdCloudCircle() {
        _classCallCheck(this, MdCloudCircle);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdCloudCircle).apply(this, arguments));
    }

    _createClass(MdCloudCircle, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm27.5 26.64q1.7166666666666686 0 2.9299999999999997-1.211666666666666t1.211666666666666-2.9299999999999997-1.211666666666666-2.9333333333333336-2.9299999999999997-1.2100000000000009h-0.8599999999999994q0-2.7333333333333343-1.9533333333333331-4.726666666666667t-4.686666666666667-1.9900000000000002q-2.344999999999999 0-4.1033333333333335 1.4450000000000003t-2.3049999999999997 3.6333333333333346l-0.2333333333333325-0.07833333333333314q-2.033333333333333 0-3.5166666666666657 1.4833333333333343t-1.4833333333333343 3.5166666666666657 1.4833333333333343 3.5166666666666657 3.5166666666666657 1.4833333333333343h14.138333333333335z m-7.5-23.28q6.875 8.881784197001252e-16 11.758333333333333 4.883333333333335t4.883333333333333 11.756666666666666-4.883333333333333 11.759999999999998-11.758333333333333 4.88333333333334-11.758333333333333-4.883333333333333-4.883333333333333-11.760000000000005 4.883333333333333-11.756666666666668 11.758333333333333-4.883333333333332z' })
                )
            );
        }
    }]);

    return MdCloudCircle;
}(React.Component);

exports.default = MdCloudCircle;
module.exports = exports['default'];