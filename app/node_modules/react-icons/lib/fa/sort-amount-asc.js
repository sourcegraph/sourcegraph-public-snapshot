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

var FaSortAmountAsc = function (_React$Component) {
    _inherits(FaSortAmountAsc, _React$Component);

    function FaSortAmountAsc() {
        _classCallCheck(this, FaSortAmountAsc);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaSortAmountAsc).apply(this, arguments));
    }

    _createClass(FaSortAmountAsc, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm16.42857142857143 32.142857142857146q0 0.2671428571428578-0.2228571428571442 0.5357142857142847l-7.119999999999999 7.121428571428567q-0.2228571428571442 0.20000000000000284-0.514285714285716 0.20000000000000284-0.2657142857142851 0-0.5114285714285707-0.20000000000000284l-7.142857142857143-7.142857142857146q-0.3342857142857143-0.3571428571428541-0.15714285714285714-0.7828571428571429 0.18285714285714283-0.44571428571427774 0.6728571428571427-0.44571428571427774h4.285714285714285v-30.714285714285715q0-0.3142857142857153 0.20000000000000018-0.5142857142857152t0.5142857142857142-0.2h4.285714285714286q0.31428571428571495 0 0.5142857142857142 0.2t0.19571428571428662 0.5142857142857142v30.714285714285715h4.2857142857142865q0.31428571428571317 0 0.5142857142857125 0.1999999999999993t0.1999999999999993 0.514285714285716z m23.57142857142857 2.857142857142854v4.285714285714285q0 0.3142857142857167-0.20000000000000284 0.5142857142857125t-0.5142857142857125 0.20000000000000284h-18.571428571428573q-0.31428571428571317 0-0.514285714285716-0.20000000000000284t-0.19999999999999574-0.5142857142857125v-4.285714285714285q0-0.3142857142857167 0.1999999999999993-0.5142857142857125t0.514285714285716-0.20000000000000284h18.571428571428577q0.3142857142857167 0 0.5142857142857125 0.20000000000000284t0.20000000000000284 0.5142857142857125z m-4.285714285714285-11.42857142857143v4.285714285714285q0 0.31428571428571317-0.20000000000000284 0.5142857142857125t-0.5142857142857125 0.1999999999999993h-14.285714285714285q-0.31428571428571317 0-0.5142857142857125-0.1999999999999993t-0.20000000000000284-0.5142857142857089v-4.285714285714285q0-0.31428571428571317 0.1999999999999993-0.514285714285716t0.514285714285716-0.1999999999999993h14.285714285714285q0.3142857142857167 0 0.5142857142857125 0.1999999999999993t0.20000000000000284 0.5142857142857125z m-4.285714285714285-11.428571428571429v4.2857142857142865q0 0.31428571428571317-0.1999999999999993 0.514285714285716t-0.514285714285716 0.1999999999999993h-10q-0.31428571428571317 0-0.5142857142857125-0.1999999999999993t-0.20000000000000284-0.5142857142857125v-4.285714285714285q0-0.31428571428571495 0.1999999999999993-0.5142857142857142t0.514285714285716-0.20000000000000284h10q0.31428571428571317 0 0.5142857142857125 0.1999999999999993t0.1999999999999993 0.5142857142857142z m-4.285714285714285-11.428571428571429v4.285714285714286q0 0.31428571428571406-0.1999999999999993 0.5142857142857142t-0.514285714285716 0.20000000000000284h-5.714285714285715q-0.31428571428571317 0-0.5142857142857125-0.20000000000000018t-0.20000000000000284-0.5142857142857142v-4.285714285714286q0-0.3142857142857143 0.1999999999999993-0.5142857142857142t0.514285714285716-0.20000000000000018h5.714285714285715q0.31428571428571317 0 0.5142857142857125 0.2t0.1999999999999993 0.5142857142857142z' })
                )
            );
        }
    }]);

    return FaSortAmountAsc;
}(React.Component);

exports.default = FaSortAmountAsc;
module.exports = exports['default'];