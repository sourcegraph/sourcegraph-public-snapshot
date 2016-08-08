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

var FaSortAlphaDesc = function (_React$Component) {
    _inherits(FaSortAlphaDesc, _React$Component);

    function FaSortAlphaDesc() {
        _classCallCheck(this, FaSortAlphaDesc);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaSortAlphaDesc).apply(this, arguments));
    }

    _createClass(FaSortAlphaDesc, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm28.014285714285712 31.964285714285715h3.9499999999999993l-1.607142857142854-4.864285714285717-0.2671428571428578-1.0500000000000007q-0.04285714285714448-0.35714285714285765-0.04285714285714448-0.4471428571428575h-0.09142857142857252l-0.06857142857142762 0.4471428571428575q0 0.022857142857141355-0.07857142857142918 0.3999999999999986t-0.16714285714285637 0.6485714285714295z m-10.157142857142858 0.1785714285714306q0 0.2671428571428578-0.2228571428571442 0.5357142857142847l-7.119999999999996 7.121428571428567q-0.22285714285714242 0.20000000000000284-0.5142857142857142 0.20000000000000284-0.2657142857142851 0-0.5114285714285707-0.20000000000000284l-7.142857142857143-7.142857142857146q-0.3342857142857145-0.3571428571428541-0.15714285714285703-0.7828571428571429 0.18285714285714283-0.44571428571427774 0.6728571428571426-0.44571428571427774h4.285714285714285v-30.714285714285715q0-0.3142857142857153 0.20000000000000018-0.5142857142857152t0.5142857142857142-0.2h4.285714285714286q0.31428571428571495 0 0.5142857142857142 0.2t0.19571428571428662 0.5142857142857142v30.714285714285715h4.285714285714285q0.31428571428571317 0 0.514285714285716 0.1999999999999993t0.1999999999999993 0.514285714285716z m20.64714285714286 5.491428571428571v2.365714285714283h-6.428571428571431v-2.365714285714283h1.6742857142857162l-1.048571428571428-3.2142857142857153h-5.4228571428571435l-1.048571428571428 3.2142857142857153h1.6742857142857126v2.365714285714283h-6.405714285714286v-2.365714285714283h1.562857142857144l5.132857142857144-14.77714285714286h3.614285714285714l5.135714285714283 14.777142857142852h1.5642857142857167z m-1.9857142857142875-25.691428571428574v5.199999999999999h-13.035714285714285v-2.008571428571429l8.237142857142857-11.80857142857143q0.2671428571428578-0.3999999999999999 0.46857142857142975-0.6028571428571428l0.24571428571428555-0.20000000000000018v-0.0685714285714285q-0.04285714285714448 0-0.14571428571428413 0.011428571428571566t-0.16714285714285637 0.011428571428571566q-0.2671428571428578 0.06714285714285717-0.6714285714285708 0.06714285714285717h-5.177142857142858v2.570000000000002h-2.677142857142858v-5.114285714285715h12.657142857142858v1.9857142857142855l-8.23857142857143 11.831428571428571q-0.13571428571428612 0.17857142857142883-0.47142857142857153 0.5800000000000001l-0.24285714285714377 0.22285714285714242v0.06714285714285673l0.3114285714285714-0.06571428571428584q0.1999999999999993-0.02285714285714313 0.6714285714285708-0.02285714285714313h5.534285714285712v-2.6571428571428566h2.700000000000003z' })
                )
            );
        }
    }]);

    return FaSortAlphaDesc;
}(React.Component);

exports.default = FaSortAlphaDesc;
module.exports = exports['default'];