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

var MdScanner = function (_React$Component) {
    _inherits(MdScanner, _React$Component);

    function MdScanner() {
        _classCallCheck(this, MdScanner);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdScanner).apply(this, arguments));
    }

    _createClass(MdScanner, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm31.640000000000004 28.36v-3.3599999999999994h-16.640000000000004v3.3599999999999994h16.64z m-20 0v-3.3599999999999994h-3.283333333333333v3.3599999999999994h3.283333333333333z m21.328333333333337-10.546666666666667q0.8616666666666646 0.23333333333333428 1.4466666666666654 1.1333333333333329t0.5849999999999937 1.913333333333334v9.14q0 1.3283333333333331-1.0166666666666657 2.3433333333333337t-2.3416666666666686 1.0166666666666657h-23.28333333333333q-1.3266666666666689 0-2.3416666666666686-1.0166666666666657t-1.0166666666666657-2.3433333333333337v-6.640000000000001q0-1.3283333333333331 1.0166666666666666-2.3433333333333337t2.3433333333333346-1.0166666666666657h20.939999999999998l-23.441666666666663-8.511666666666667 1.171666666666666-3.125z' })
                )
            );
        }
    }]);

    return MdScanner;
}(React.Component);

exports.default = MdScanner;
module.exports = exports['default'];