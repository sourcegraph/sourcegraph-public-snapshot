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

var FaMailReply = function (_React$Component) {
    _inherits(FaMailReply, _React$Component);

    function FaMailReply() {
        _classCallCheck(this, FaMailReply);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaMailReply).apply(this, arguments));
    }

    _createClass(FaMailReply, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm40 25q0 3.7057142857142864-2.8342857142857127 10.067142857142855-0.06714285714285495 0.15714285714285836-0.23428571428571132 0.5357142857142847t-0.29999999999999716 0.6714285714285708-0.29142857142856826 0.490000000000002q-0.2685714285714411 0.37857142857143344-0.6257142857142952 0.37857142857143344-0.33428571428571274 0-0.5242857142857176-0.22142857142856798t-0.18999999999999773-0.5571428571428569q0-0.20285714285714107 0.05714285714285694-0.5928571428571416t0.054285714285711606-0.5242857142857176q0.11142857142856855-1.5171428571428578 0.11142857142856855-2.7457142857142856 0-2.252857142857142-0.39000000000000057-4.03857142857143t-1.0828571428571436-3.0914285714285725-1.7857142857142847-2.2571428571428562-2.354285714285716-1.5500000000000007-2.9685714285714297-0.9485714285714302-3.4385714285714286-0.47857142857142776-3.9171428571428493-0.1371428571428588h-4.999999999999998v5.714285714285715q0 0.581428571428571-0.4228571428571435 1.0057142857142871t-1.0057142857142853 0.42285714285713993-1.0028571428571436-0.4228571428571435l-11.428571428571429-11.428571428571429q-0.42571428571428527-0.4242857142857126-0.42571428571428527-1.0057142857142836t0.42857142857142855-1l11.428571428571429-11.428571428571429q0.4214285714285726-0.42857142857142905 1-0.42857142857142905t1.0057142857142853 0.4257142857142857 0.4228571428571435 1.002857142857143v5.7142857142857135h4.999999999999998q15.918571428571433 0 19.534285714285716 9.000000000000002 1.1799999999999997 2.9885714285714258 1.1799999999999997 7.428571428571427z' })
                )
            );
        }
    }]);

    return FaMailReply;
}(React.Component);

exports.default = FaMailReply;
module.exports = exports['default'];