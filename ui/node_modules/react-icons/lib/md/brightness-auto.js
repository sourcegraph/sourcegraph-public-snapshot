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

var MdBrightnessAuto = function (_React$Component) {
    _inherits(MdBrightnessAuto, _React$Component);

    function MdBrightnessAuto() {
        _classCallCheck(this, MdBrightnessAuto);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdBrightnessAuto).apply(this, arguments));
    }

    _createClass(MdBrightnessAuto, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm23.828333333333337 26.64h3.203333333333333l-5.391666666666666-15h-3.283333333333335l-5.388333333333334 15h3.2033333333333314l1.1716666666666669-3.2833333333333314h5.313333333333333z m9.533333333333335-12.186666666666667l5.464999999999996 5.546666666666667-5.466666666666669 5.546666666666667v7.813333333333333h-7.816666666666666l-5.543333333333333 5.466666666666669-5.546666666666667-5.466666666666669h-7.813333333333333v-7.813333333333333l-5.466666666666667-5.546666666666667 5.466666666666667-5.546666666666667v-7.813333333333333h7.813333333333333l5.546666666666667-5.466666666666667 5.546666666666667 5.466666666666667h7.813333333333333v7.813333333333333z m-15.316666666666666 6.640000000000001l1.9549999999999947-6.093333333333334 1.9533333333333331 6.093333333333334h-3.9066666666666663z' })
                )
            );
        }
    }]);

    return MdBrightnessAuto;
}(React.Component);

exports.default = MdBrightnessAuto;
module.exports = exports['default'];