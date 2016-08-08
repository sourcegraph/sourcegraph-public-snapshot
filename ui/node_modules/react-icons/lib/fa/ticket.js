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

var FaTicket = function (_React$Component) {
    _inherits(FaTicket, _React$Component);

    function FaTicket() {
        _classCallCheck(this, FaTicket);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaTicket).apply(this, arguments));
    }

    _createClass(FaTicket, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm22.857142857142858 10.09l7.052857142857146 7.0528571428571425-12.767142857142861 12.767142857142861-7.0528571428571425-7.052857142857146z m-4.710000000000001 21.852857142857143l13.794285714285717-13.795714285714286q0.42285714285713993-0.4242857142857126 0.42285714285713993-1.0042857142857144t-0.4242857142857126-1.0042857142857144l-8.079999999999998-8.081428571428571q-0.3999999999999986-0.40000000000000036-1.0042857142857144-0.40000000000000036t-1.0042857142857144 0.40000000000000036l-13.794285714285717 13.797142857142859q-0.4228571428571426 0.4242857142857126-0.4228571428571426 1.0042857142857144t0.4242857142857135 1.0042857142857144l8.081428571428573 8.079999999999998q0.3999999999999986 0.3999999999999986 1.0042857142857144 0.3999999999999986t1.0042857142857144-0.3999999999999986z m19.842857142857145-14.219999999999999l-20.242857142857144 20.267142857142858q-0.8285714285714292 0.8257142857142838-2.0214285714285722 0.8257142857142838t-2.0199999999999996-0.8242857142857147l-2.814285714285715-2.8142857142857167q1.251428571428571-1.2485714285714238 1.251428571428571-3.0342857142857085t-1.25-3.0371428571428574-3.0357142857142847-1.2485714285714309-3.0357142857142856 1.2485714285714273l-2.79-2.814285714285713q-0.8257142857142856-0.8242857142857147-0.8257142857142856-2.0185714285714305t0.8257142857142856-2.018571428571427l20.245714285714286-20.225714285714286q0.8257142857142838-0.8242857142857143 2.0199999999999996-0.8242857142857143t2.0199999999999996 0.8257142857142856l2.789999999999999 2.7914285714285714q-1.2499999999999964 1.25-1.2499999999999964 3.0357142857142865t1.25 3.0342857142857156 3.0357142857142883 1.2514285714285691 3.0357142857142847-1.251428571428571l2.8142857142857167 2.790000000000001q0.8242857142857147 0.8257142857142856 0.8242857142857147 2.0199999999999996t-0.8257142857142838 2.0214285714285722z' })
                )
            );
        }
    }]);

    return FaTicket;
}(React.Component);

exports.default = FaTicket;
module.exports = exports['default'];