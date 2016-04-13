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

var MdSettingsEthernet = function (_React$Component) {
    _inherits(MdSettingsEthernet, _React$Component);

    function MdSettingsEthernet() {
        _classCallCheck(this, MdSettingsEthernet);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdSettingsEthernet).apply(this, arguments));
    }

    _createClass(MdSettingsEthernet, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm29.61 9.14l9.06333333333334 10.86-9.063333333333333 10.86-2.578333333333333-2.1099999999999994 7.263333333333328-8.75-7.266666666666666-8.75z m-11.25 12.5v-3.2833333333333314h3.2833333333333314v3.2833333333333314h-3.2833333333333314z m10-3.280000000000001v3.2833333333333314h-3.3599999999999994v-3.2833333333333314h3.3599999999999994z m-16.72 3.280000000000001v-3.2833333333333314h3.3599999999999994v3.2833333333333314h-3.3599999999999994z m1.3266666666666662-10.39l-7.261666666666667 8.75 7.2666666666666675 8.75-2.58 2.1099999999999994-9.065000000000001-10.86 9.063333333333333-10.860000000000001z' })
                )
            );
        }
    }]);

    return MdSettingsEthernet;
}(React.Component);

exports.default = MdSettingsEthernet;
module.exports = exports['default'];