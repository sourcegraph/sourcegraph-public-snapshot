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

var MdShuffle = function (_React$Component) {
    _inherits(MdShuffle, _React$Component);

    function MdShuffle() {
        _classCallCheck(this, MdShuffle);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdShuffle).apply(this, arguments));
    }

    _createClass(MdShuffle, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm24.688333333333336 22.343333333333334l5.233333333333334 5.233333333333334 3.4400000000000013-3.4366666666666674v9.216666666666669h-9.216666666666669l3.4366666666666674-3.4366666666666674-5.236666666666665-5.233333333333334z m-0.5500000000000007-15.703333333333335h9.219999999999999v9.216666666666667l-3.4383333333333326-3.4366666666666674-20.936666666666667 20.938333333333336-2.3466666666666667-2.3416666666666686 20.94-20.941666666666663z m-6.483333333333334 8.673333333333334l-2.3383333333333347 2.3433333333333337-8.675-8.673333333333334 2.341666666666667-2.339999999999999z' })
                )
            );
        }
    }]);

    return MdShuffle;
}(React.Component);

exports.default = MdShuffle;
module.exports = exports['default'];