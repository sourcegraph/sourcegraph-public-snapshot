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

var MdQueue = function (_React$Component) {
    _inherits(MdQueue, _React$Component);

    function MdQueue() {
        _classCallCheck(this, MdQueue);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdQueue).apply(this, arguments));
    }

    _createClass(MdQueue, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm31.640000000000004 18.36v-3.3599999999999994h-6.640000000000004v-6.639999999999999h-3.3599999999999994v6.639999999999999h-6.640000000000001v3.3599999999999994h6.640000000000001v6.640000000000001h3.3599999999999994v-6.640000000000001h6.640000000000001z m1.7200000000000024-15q1.3283333333333331 0 2.3049999999999997 0.9766666666666666t0.9750000000000014 2.3049999999999997v20q0 1.3283333333333331-0.9766666666666666 2.3433333333333337t-2.306666666666665 1.0166666666666657h-20q-1.3283333333333331 0-2.3433333333333337-1.0166666666666657t-1.0133333333333425-2.341666666666665v-20q0-1.3283333333333331 1.0166666666666657-2.3049999999999997t2.3433333333333337-0.9766666666666666h20z m-26.720000000000006 6.640000000000001v23.36h23.36v3.2833333333333314h-23.36q-1.3283333333333331 0-2.3049999999999997-0.9783333333333317t-0.9750000000000001-2.306666666666665v-23.358333333333334h3.2833333333333337z' })
                )
            );
        }
    }]);

    return MdQueue;
}(React.Component);

exports.default = MdQueue;
module.exports = exports['default'];