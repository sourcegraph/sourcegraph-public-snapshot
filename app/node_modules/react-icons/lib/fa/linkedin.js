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

var FaLinkedin = function (_React$Component) {
    _inherits(FaLinkedin, _React$Component);

    function FaLinkedin() {
        _classCallCheck(this, FaLinkedin);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaLinkedin).apply(this, arguments));
    }

    _createClass(FaLinkedin, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm10.647142857142859 13.951428571428572v22.119999999999997h-7.3657142857142865v-22.119999999999997h7.365714285714285z m0.468571428571428-6.831428571428572q0.024285714285714022 1.628571428571429-1.1257142857142863 2.7228571428571433t-3.024285714285714 1.0942857142857143h-0.042857142857142705q-1.831428571428571 0-2.948571428571429-1.0942857142857143t-1.1171428571428574-2.7228571428571433q0-1.6514285714285712 1.15-2.734285714285715t3.0028571428571427-1.0828571428571427 2.968571428571428 1.0828571428571427 1.138571428571428 2.734285714285714z m26.02714285714286 16.27142857142857v12.68h-7.342857142857142v-11.82857142857143q0-2.345714285714287-0.9057142857142857-3.671428571428571t-2.8242857142857147-1.3285714285714292q-1.4071428571428584 0-2.3571428571428577 0.7685714285714269t-1.4142857142857146 1.9085714285714275q-0.24714285714285822 0.6714285714285708-0.24714285714285822 1.8085714285714296v12.342857142857145h-7.342857142857142q0.042857142857142705-8.904285714285713 0.042857142857142705-14.440000000000001t-0.02285714285714313-6.607142857142858l-0.02142857142857224-1.0714285714285712h7.342857142857142v3.2142857142857135h-0.04285714285714448q0.4471428571428575-0.7142857142857153 0.9142857142857146-1.25t1.2628571428571433-1.1600000000000001 1.9400000000000013-0.9714285714285715 2.557142857142857-0.3457142857142852q3.814285714285713 0 6.137142857142859 2.532857142857143t2.3242857142857147 7.422857142857145z' })
                )
            );
        }
    }]);

    return FaLinkedin;
}(React.Component);

exports.default = FaLinkedin;
module.exports = exports['default'];