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

var MdExplore = function (_React$Component) {
    _inherits(MdExplore, _React$Component);

    function MdExplore() {
        _classCallCheck(this, MdExplore);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdExplore).apply(this, arguments));
    }

    _createClass(MdExplore, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm23.671666666666667 23.671666666666667l6.328333333333333-13.671666666666667-13.671666666666663 6.328333333333337-6.328333333333337 13.671666666666663z m-3.671666666666667-20.311666666666667q6.875 8.881784197001252e-16 11.758333333333333 4.883333333333335t4.883333333333333 11.756666666666666-4.883333333333333 11.759999999999998-11.758333333333333 4.88333333333334-11.758333333333333-4.883333333333333-4.883333333333333-11.760000000000005 4.883333333333333-11.756666666666668 11.758333333333333-4.883333333333332z m0 14.843333333333337q0.783333333333335 0 1.288333333333334 0.5083333333333329t0.5083333333333329 1.288333333333334-0.5083333333333329 1.288333333333334-1.288333333333334 0.5083333333333329-1.288333333333334-0.5083333333333329-0.5083333333333329-1.288333333333334 0.5083333333333329-1.288333333333334 1.288333333333334-0.5083333333333329z' })
                )
            );
        }
    }]);

    return MdExplore;
}(React.Component);

exports.default = MdExplore;
module.exports = exports['default'];