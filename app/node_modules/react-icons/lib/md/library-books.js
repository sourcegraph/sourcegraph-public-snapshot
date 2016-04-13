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

var MdLibraryBooks = function (_React$Component) {
    _inherits(MdLibraryBooks, _React$Component);

    function MdLibraryBooks() {
        _classCallCheck(this, MdLibraryBooks);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdLibraryBooks).apply(this, arguments));
    }

    _createClass(MdLibraryBooks, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm31.640000000000004 11.64v-3.283333333333333h-16.640000000000004v3.283333333333333h16.64z m-6.640000000000004 13.36v-3.3599999999999994h-10v3.3599999999999994h10z m6.640000000000001-6.640000000000001v-3.3599999999999994h-16.64v3.3599999999999994h16.64z m1.7199999999999989-15q1.3283333333333331 0 2.3049999999999997 0.9766666666666666t0.9750000000000014 2.3049999999999997v20q0 1.3283333333333331-0.9766666666666666 2.3433333333333337t-2.306666666666665 1.0166666666666657h-20q-1.3283333333333331 0-2.3433333333333337-1.0166666666666657t-1.0133333333333354-2.341666666666665v-20q0-1.3283333333333331 1.0166666666666657-2.3049999999999997t2.3433333333333337-0.9766666666666666h20z m-26.72 6.640000000000001v23.36h23.36v3.2833333333333314h-23.36q-1.3283333333333331 0-2.3049999999999997-0.9783333333333317t-0.9750000000000001-2.306666666666665v-23.358333333333334h3.2833333333333337z' })
                )
            );
        }
    }]);

    return MdLibraryBooks;
}(React.Component);

exports.default = MdLibraryBooks;
module.exports = exports['default'];