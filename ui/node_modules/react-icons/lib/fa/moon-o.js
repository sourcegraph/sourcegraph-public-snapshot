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

var FaMoonO = function (_React$Component) {
    _inherits(FaMoonO, _React$Component);

    function FaMoonO() {
        _classCallCheck(this, FaMoonO);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaMoonO).apply(this, arguments));
    }

    _createClass(FaMoonO, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm31.74285714285714 29.085714285714285q-1.2071428571428555 0.1999999999999993-2.4571428571428555 0.1999999999999993-4.062857142857144 0-7.522857142857141-2.008571428571429t-5.46857142857143-5.46857142857143-2.008571428571427-7.52285714285714q0-4.2857142857142865 2.321428571428571-7.968571428571429-4.4857142857142875 1.3399999999999999-7.332857142857144 5.111428571428571t-2.845714285714285 8.571428571428571q0 2.8999999999999986 1.1385714285714288 5.547142857142859t3.0471428571428563 4.5528571428571425 4.5528571428571425 3.047142857142859 5.547142857142859 1.1385714285714243q3.2142857142857153 0 6.104285714285716-1.3714285714285737t4.921428571428571-3.828571428571429z m4.528571428571432-1.8999999999999986q-2.097142857142856 4.532857142857143-6.328571428571429 7.242857142857144t-9.228571428571428 2.7142857142857153q-3.482857142857142 0-6.651428571428571-1.3628571428571448t-5.46857142857143-3.6599999999999966-3.66-5.46857142857143-1.3628571428571425-6.651428571428575q0-3.4142857142857146 1.285714285714286-6.5285714285714285t3.4814285714285713-5.390000000000001 5.257142857142858-3.671428571428571 6.4714285714285715-1.5285714285714285q0.985714285714284-0.042857142857142705 1.3628571428571412 0.8714285714285714 0.4028571428571439 0.9142857142857141-0.33285714285714363 1.6057142857142854-1.9200000000000017 1.7428571428571429-2.935714285714287 4.052857142857144t-1.0157142857142816 4.874285714285714q0 3.3042857142857134 1.62857142857143 6.095714285714285t4.420000000000002 4.420000000000002 6.094285714285714 1.6314285714285717q2.6342857142857135 0 5.09-1.138571428571428 0.914285714285711-0.3999999999999986 1.607142857142854 0.28999999999999915 0.3142857142857167 0.31428571428571317 0.39000000000000057 0.7571428571428562t-0.10000000000000142 0.8500000000000014z' })
                )
            );
        }
    }]);

    return FaMoonO;
}(React.Component);

exports.default = FaMoonO;
module.exports = exports['default'];