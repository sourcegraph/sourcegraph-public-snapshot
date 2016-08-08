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

var MdSyncDisabled = function (_React$Component) {
    _inherits(MdSyncDisabled, _React$Component);

    function MdSyncDisabled() {
        _classCallCheck(this, MdSyncDisabled);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdSyncDisabled).apply(this, arguments));
    }

    _createClass(MdSyncDisabled, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm33.36 6.640000000000001l-3.9833333333333343 3.9833333333333343q3.9833333333333343 3.9866666666666664 3.9833333333333343 9.376666666666665 0 3.75-2.0333333333333314 7.033333333333331l-2.5-2.423333333333332q1.173333333333332-2.3433333333333337 1.173333333333332-4.609999999999999 0-4.063333333333333-2.9666666666666686-7.033333333333333l-3.673333333333332 3.673333333333334v-10h10z m-28.593333333333334 2.343333333333332l2.1083333333333343-2.1083333333333334 26.171666666666667 26.25-2.1099999999999994 2.1099999999999994-3.9066666666666663-3.905000000000001q-1.875 1.0933333333333337-3.75 1.5633333333333326v-3.4383333333333326q0.625-0.23333333333333428 1.3283333333333331-0.625l-13.438333333333334-13.439999999999998q-1.17 2.3433333333333337-1.17 4.609999999999999 0 4.063333333333333 2.966666666666667 7.033333333333331l3.671666666666665-3.673333333333332v10h-10l3.9833333333333343-3.9833333333333343q-3.9816666666666656-3.9883333333333297-3.9816666666666656-9.376666666666665 0-3.75 2.033333333333333-7.033333333333333z m11.873333333333335 1.5633333333333326q-0.466666666666665 0.1566666666666663-1.25 0.625l-2.421666666666667-2.5q1.955-1.1716666666666669 3.673333333333332-1.5633333333333335v3.4383333333333335z' })
                )
            );
        }
    }]);

    return MdSyncDisabled;
}(React.Component);

exports.default = MdSyncDisabled;
module.exports = exports['default'];