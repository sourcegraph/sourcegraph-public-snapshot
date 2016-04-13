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

var MdSmokeFree = function (_React$Component) {
    _inherits(MdSmokeFree, _React$Component);

    function MdSmokeFree() {
        _classCallCheck(this, MdSmokeFree);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdSmokeFree).apply(this, arguments));
    }

    _createClass(MdSmokeFree, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm28.36 26.563333333333336l-4.921666666666667-4.921666666666667h4.921666666666667v4.921666666666667z m-4.219999999999999-12.033333333333333q-2.2666666666666657 0-3.9066666666666663-1.6783333333333328t-1.6400000000000006-3.9433333333333334 1.6400000000000006-3.9083333333333368 3.9066666666666663-1.6400000000000001v2.4999999999999996q-1.3283333333333331 0-2.1883333333333326 0.8200000000000003t-0.8583333333333343 2.0700000000000003q0 1.3283333333333331 0.8999999999999986 2.3433333333333337t2.1466666666666683 1.0166666666666657h2.578333333333333q2.3433333333333337 0 4.063333333333333 1.5216666666666665t1.7166666666666686 3.709999999999999v2.658333333333335h-2.498333333333335v-2.1099999999999994q0-1.5633333333333326-0.9750000000000014-2.461666666666666t-2.3049999999999997-0.9000000000000004h-2.578333333333333z m7.266666666666666-6.404999999999999q2.3433333333333337 1.0933333333333337 3.788333333333334 3.3599999999999994t1.4450000000000003 5.080000000000002v3.434999999999995h-2.5v-3.4333333333333336q0-2.8133333333333326-1.913333333333334-4.805t-4.726666666666667-1.9949999999999992v-2.5q1.25 0 2.1499999999999986-0.9000000000000004t0.8966666666666683-2.2249999999999996h2.5q0 2.3433333333333337-1.6400000000000006 3.9833333333333343z m-1.4066666666666663 13.516666666666662h2.5v5h-2.5v-5z m4.140000000000001 0h2.5v5h-2.5v-5z m-30.78-11.641666666666666l2.033333333333334-2.1100000000000003 28.356666666666666 28.36-2.1083333333333343 2.1099999999999994-11.641666666666666-11.716666666666665h-16.636666666666667v-5h11.636666666666667z' })
                )
            );
        }
    }]);

    return MdSmokeFree;
}(React.Component);

exports.default = MdSmokeFree;
module.exports = exports['default'];