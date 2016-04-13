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

var FaItalic = function (_React$Component) {
    _inherits(FaItalic, _React$Component);

    function FaItalic() {
        _classCallCheck(this, FaItalic);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaItalic).apply(this, arguments));
    }

    _createClass(FaItalic, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm8.571428571428571 37.1l0.3800000000000008-1.8999999999999986q0.13428571428571345-0.04285714285714448 1.8185714285714276-0.47857142857142776t2.4857142857142858-0.8371428571428581q0.6285714285714281-0.7828571428571465 0.9171428571428564-2.2571428571428562 0.02285714285714313-0.15428571428571303 1.3857142857142861-6.449999999999999t2.5428571428571427-12.131428571428573 1.1600000000000001-6.618571428571428v-0.5571428571428569q-0.5357142857142847-0.2914285714285718-1.217142857142857-0.4142857142857146t-1.5514285714285734-0.17857142857142883-1.2942857142857136-0.12285714285714278l0.4257142857142888-2.2971428571428594q0.7371428571428567 0.042857142857142705 2.678571428571427 0.1457142857142859t3.337142857142858 0.15714285714285703 2.6900000000000013 0.05428571428571427q1.071428571428573 0 2.1999999999999993-0.05714285714285694t2.6999999999999993-0.15428571428571436 2.19857142857143-0.1457142857142859q-0.1114285714285721 0.8714285714285714-0.4242857142857126 1.9857142857142853-0.6714285714285708 0.2242857142857142-2.265714285714285 0.637142857142857t-2.421428571428571 0.7471428571428573q-0.17857142857142705 0.4228571428571426-0.31428571428571317 0.9471428571428575t-0.1999999999999993 0.8928571428571423-0.16714285714285637 1.0142857142857142-0.14571428571428413 0.9385714285714286q-0.6028571428571432 3.3028571428571425-1.952857142857141 9.364285714285714t-1.7285714285714278 7.935714285714287q-0.04571428571428626 0.1999999999999993-0.2914285714285718 1.2942857142857136t-0.4471428571428575 2.0100000000000016-0.35714285714285765 1.8642857142857174-0.13571428571428612 1.2828571428571394l0.022857142857141355 0.3999999999999986q0.38142857142857167 0.09142857142857252 4.131428571428572 0.6928571428571431-0.0671428571428585 0.9828571428571422-0.35714285714285765 2.210000000000001-0.24571428571428555 0-0.725714285714286 0.032857142857139365t-0.725714285714286 0.032857142857139365q-0.6471428571428568 0-1.942857142857143-0.2228571428571442t-1.918571428571429-0.2228571428571442q-3.08-0.04285714285714448-4.600000000000001-0.04285714285714448-1.137142857142857 0-3.1900000000000013 0.20000000000000284t-2.7000000000000046 0.24571428571429266z' })
                )
            );
        }
    }]);

    return FaItalic;
}(React.Component);

exports.default = FaItalic;
module.exports = exports['default'];