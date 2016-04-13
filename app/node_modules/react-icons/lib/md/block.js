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

var MdBlock = function (_React$Component) {
    _inherits(MdBlock, _React$Component);

    function MdBlock() {
        _classCallCheck(this, MdBlock);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdBlock).apply(this, arguments));
    }

    _createClass(MdBlock, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm20 33.36q5.466666666666669 0 9.413333333333334-3.9450000000000003t3.9450000000000003-9.415q0-4.686666666666667-2.8133333333333326-8.200000000000001l-18.75 18.75q3.5166666666666657 2.8116666666666674 8.203333333333333 2.8116666666666674z m-13.36-13.36q0 4.688333333333333 2.8133333333333326 8.203333333333333l18.75-18.75q-3.5166666666666657-2.8133333333333326-8.203333333333333-2.8133333333333326-5.466666666666667 0-9.413333333333334 3.9450000000000003t-3.9449999999999994 9.415z m13.36-16.64q6.875 8.881784197001252e-16 11.758333333333333 4.883333333333335t4.883333333333333 11.756666666666666-4.883333333333333 11.759999999999998-11.758333333333333 4.88333333333334-11.758333333333333-4.883333333333333-4.883333333333333-11.760000000000005 4.883333333333333-11.756666666666668 11.758333333333333-4.883333333333332z' })
                )
            );
        }
    }]);

    return MdBlock;
}(React.Component);

exports.default = MdBlock;
module.exports = exports['default'];