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

var FaMercury = function (_React$Component) {
    _inherits(FaMercury, _React$Component);

    function FaMercury() {
        _classCallCheck(this, FaMercury);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaMercury).apply(this, arguments));
    }

    _createClass(FaMercury, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm25.67142857142857 7.052857142857143q3.2357142857142875 1.6071428571428568 5.210000000000001 4.7t1.9757142857142895 6.815714285714284q0 4.931428571428572-3.2928571428571445 8.581428571428571t-8.135714285714286 4.188571428571432v2.947142857142854h2.1428571428571423q0.31428571428571317 0 0.514285714285716 0.20000000000000284t0.1999999999999993 0.5142857142857125v1.4285714285714306q0 0.3142857142857167-0.1999999999999993 0.5142857142857125t-0.5142857142857125 0.20000000000000284h-2.142857142857146v2.142857142857146q0 0.3142857142857167-0.1999999999999993 0.5142857142857125t-0.514285714285716 0.19999999999999574h-1.428571428571427q-0.31428571428571317 0-0.514285714285716-0.20000000000000284t-0.1999999999999993-0.5142857142857125v-2.142857142857139h-2.1428571428571423q-0.31428571428571317 0-0.5142857142857142-0.20000000000000284t-0.20000000000000107-0.5142857142857125v-1.4285714285714306q0-0.3142857142857167 0.1999999999999993-0.5142857142857125t0.514285714285716-0.20000000000000284h2.1428571428571423v-2.9471428571428575q-4.842857142857143-0.5357142857142847-8.135714285714286-4.185714285714287t-3.2928571428571436-8.581428571428567q0-3.7285714285714278 1.975714285714285-6.82t5.211428571428572-4.7q-3.6814285714285706-2.1414285714285723-5.087142857142858-6.0914285714285725-0.13428571428571345-0.357142857142857 0.07857142857142918-0.657142857142857t0.588571428571429-0.30285714285714294h1.5399999999999991q0.468571428571428 0 0.6471428571428568 0.44571428571428573 0.9828571428571422 2.3657142857142857 3.1257142857142863 3.8171428571428576t4.777142857142858 1.451428571428571 4.777142857142856-1.451428571428571 3.125714285714288-3.8171428571428576q0.17857142857142705-0.44571428571428573 0.8257142857142838-0.44571428571428573h1.361428571428572q0.379999999999999 0 0.5914285714285725 0.3t0.07857142857142918 0.657142857142857q-1.4057142857142857 3.952857142857143-5.09 6.095714285714286z m-5.671428571428571 21.51857142857143q4.12857142857143 0 7.064285714285717-2.935714285714287t2.9357142857142833-7.064285714285713-2.935714285714287-7.064285714285715-7.064285714285713-2.935714285714287-7.064285714285715 2.935714285714285-2.935714285714285 7.064285714285717 2.935714285714287 7.064285714285713 7.064285714285713 2.935714285714287z' })
                )
            );
        }
    }]);

    return FaMercury;
}(React.Component);

exports.default = FaMercury;
module.exports = exports['default'];