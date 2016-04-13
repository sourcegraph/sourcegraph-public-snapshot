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

var FaMarsStrokeV = function (_React$Component) {
    _inherits(FaMarsStrokeV, _React$Component);

    function FaMarsStrokeV() {
        _classCallCheck(this, FaMarsStrokeV);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaMarsStrokeV).apply(this, arguments));
    }

    _createClass(FaMarsStrokeV, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm21.42857142857143 14.375714285714286q4.842857142857142 0.5357142857142865 8.135714285714286 4.185714285714285t3.2928571428571445 8.581428571428571q0 3.7285714285714278-1.942857142857143 6.831428571428571t-5.265714285714285 4.732857142857142-7.120000000000001 1.2057142857142864q-2.968571428571428-0.33571428571428896-5.48-1.9657142857142844t-4.062857142857146-4.195714285714281-1.7971428571428572-5.557142857142857q-0.2671428571428569-3.46142857142857 1.1714285714285708-6.518571428571427t4.151428571428571-5 6.060000000000002-2.300000000000006v-2.944285714285714h-3.571428571428573q-0.31428571428571495 0-0.5142857142857142-0.1999999999999993t-0.1999999999999993-0.5171428571428578v-1.4285714285714288q0-0.3114285714285714 0.1999999999999993-0.5114285714285707t0.5142857142857142-0.1999999999999993h3.571428571428573v-3.6885714285714304l-2.052857142857146 2.057142857142857q-0.2228571428571442 0.20000000000000018-0.5142857142857125 0.20000000000000018t-0.4900000000000002-0.20000000000000018l-1.0285714285714285-1.0285714285714285q-0.1999999999999993-0.20000000000000018-0.1999999999999993-0.48857142857142843t0.1999999999999993-0.5142857142857142l4.5114285714285725-4.485714285714286q0.42285714285713993-0.42571428571428616 1.0028571428571418-0.42571428571428616t1.0042857142857144 0.42428571428571427l4.508571428571429 4.485714285714286q0.1999999999999993 0.2242857142857142 0.1999999999999993 0.5142857142857142t-0.1999999999999993 0.4914285714285711l-1.0285714285714285 1.0285714285714285q-0.19857142857143018 0.1985714285714293-0.490000000000002 0.1985714285714293t-0.514285714285716-0.20000000000000018l-2.0514285714285663-2.0528571428571425v3.6814285714285706h3.571428571428573q0.31428571428571317 0 0.514285714285716 0.20285714285714285t0.1999999999999993 0.5142857142857142v1.4285714285714288q0 0.3114285714285714-0.1999999999999993 0.5114285714285707t-0.5142857142857196 0.20000000000000107h-3.5714285714285694v2.9485714285714284z m-1.4285714285714306 22.767142857142858q4.12857142857143 0 7.064285714285717-2.9357142857142833t2.9357142857142833-7.06428571428572-2.935714285714287-7.064285714285713-7.064285714285713-2.935714285714287-7.064285714285715 2.935714285714287-2.935714285714285 7.064285714285713 2.935714285714287 7.064285714285713 7.064285714285713 2.9357142857142904z' })
                )
            );
        }
    }]);

    return FaMarsStrokeV;
}(React.Component);

exports.default = FaMarsStrokeV;
module.exports = exports['default'];