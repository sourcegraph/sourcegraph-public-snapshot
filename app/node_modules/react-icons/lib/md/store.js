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

var MdStore = function (_React$Component) {
    _inherits(MdStore, _React$Component);

    function MdStore() {
        _classCallCheck(this, MdStore);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdStore).apply(this, arguments));
    }

    _createClass(MdStore, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm20 30v-6.640000000000001h-10v6.640000000000001h10z m15-6.640000000000001h-1.6400000000000006v10h-3.3599999999999994v-10h-6.640000000000001v10h-16.716666666666665v-10h-1.6433333333333344v-3.3599999999999994l1.6416666666666666-8.360000000000001h26.71666666666667l1.6416666666666657 8.360000000000001v3.3599999999999994z m-1.6400000000000006-16.720000000000002v3.360000000000003h-26.716666666666665v-3.3599999999999994h26.716666666666665z' })
                )
            );
        }
    }]);

    return MdStore;
}(React.Component);

exports.default = MdStore;
module.exports = exports['default'];