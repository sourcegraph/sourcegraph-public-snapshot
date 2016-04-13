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

var MdAddAlert = function (_React$Component) {
    _inherits(MdAddAlert, _React$Component);

    function MdAddAlert() {
        _classCallCheck(this, MdAddAlert);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdAddAlert).apply(this, arguments));
    }

    _createClass(MdAddAlert, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm26.64 21.716666666666665v-3.3583333333333343h-5v-5h-3.2833333333333314v5h-5v3.3599999999999994h5v5h3.2833333333333314v-5h5z m4.844999999999999 6.330000000000002l3.5166666666666657 3.5166666666666657v1.7950000000000017h-30.001666666666665v-1.7966666666666669l3.5166666666666657-3.5166666666666657v-9.686666666666667q0-3.9833333333333343 2.5-7.15t6.326666666666668-4.0216666666666665v-1.1700000000000008q0-1.0949999999999998 0.783333333333335-1.876666666666666t1.8733333333333313-0.7833333333333341 1.875 0.7833333333333332 0.783333333333335 1.875v1.17q3.826666666666668 0.8600000000000003 6.326666666666664 4.0233333333333325t2.5 7.150000000000002v9.686666666666667z m-14.768333333333334 6.953333333333333h6.566666666666666q0 1.3283333333333331-0.9783333333333317 2.3433333333333337t-2.3049999999999997 1.0166666666666657-2.3049999999999997-1.0166666666666657-0.9783333333333353-2.3433333333333337z' })
                )
            );
        }
    }]);

    return MdAddAlert;
}(React.Component);

exports.default = MdAddAlert;
module.exports = exports['default'];