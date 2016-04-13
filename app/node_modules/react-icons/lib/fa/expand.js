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

var FaExpand = function (_React$Component) {
    _inherits(FaExpand, _React$Component);

    function FaExpand() {
        _classCallCheck(this, FaExpand);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaExpand).apply(this, arguments));
    }

    _createClass(FaExpand, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm19.71 23.571428571428573q0 0.28999999999999915-0.2228571428571442 0.514285714285716l-7.4085714285714275 7.4085714285714275 3.2142857142857153 3.2142857142857153q0.4214285714285708 0.42428571428570905 0.4214285714285708 1.0057142857142836t-0.4228571428571435 1.0028571428571453-1.0057142857142853 0.42571428571428527h-10q-0.5771428571428578 0-1.0000000000000009-0.4285714285714306t-0.4285714285714284-1v-10q0-0.581428571428571 0.4257142857142857-1.0057142857142871t1.0028571428571427-0.4228571428571435 1.0057142857142862 0.4228571428571435l3.2142857142857153 3.2142857142857153 7.4085714285714275-7.408571428571431q0.2242857142857133-0.2242857142857133 0.514285714285716-0.2242857142857133t0.5142857142857125 0.2228571428571442l2.5428571428571445 2.5428571428571445q0.2228571428571442 0.2228571428571442 0.2228571428571442 0.5142857142857125z m17.432857142857145-19.28571428571429v10q0 0.5800000000000001-0.42428571428571615 1.0042857142857144t-1.0042857142857144 0.42428571428571615-1.0042857142857144-0.4242857142857144l-3.2142857142857153-3.2142857142857135-7.41 7.41q-0.2228571428571442 0.2228571428571442-0.514285714285716 0.2228571428571442t-0.5114285714285707-0.2228571428571442l-2.5428571428571445-2.5428571428571445q-0.2242857142857133-0.2242857142857133-0.2242857142857133-0.5142857142857125t0.2228571428571442-0.5142857142857142l7.411428571428569-7.408571428571429-3.2142857142857153-3.2142857142857144q-0.4271428571428544-0.42428571428571527-0.4271428571428544-1.0057142857142871t0.4285714285714306-1 1-0.4285714285714284h10q0.5814285714285745 0 1.0057142857142836 0.4257142857142857t0.42285714285714704 1.0028571428571427z' })
                )
            );
        }
    }]);

    return FaExpand;
}(React.Component);

exports.default = FaExpand;
module.exports = exports['default'];