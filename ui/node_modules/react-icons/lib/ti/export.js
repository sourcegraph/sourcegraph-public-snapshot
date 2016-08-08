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

var TiExport = function (_React$Component) {
    _inherits(TiExport, _React$Component);

    function TiExport() {
        _classCallCheck(this, TiExport);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(TiExport).apply(this, arguments));
    }

    _createClass(TiExport, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm13.333333333333334 27.5v0.8333333333333321c2.816666666666668-4.296666666666667 6.000000000000002-6.588333333333335 10.000000000000002-6.666666666666668v5c0 0.9166666666666679 0.8500000000000014 1.6666666666666679 1.9050000000000011 1.6666666666666679 0.6066666666666656 0 1.125-0.26333333333333186 1.4716666666666676-0.6499999999999986 3.2233333333333327-3.383333333333333 9.956666666666667-10.183333333333334 9.956666666666667-10.183333333333334s-6.733333333333334-6.803333333333335-9.958333333333336-10.228333333333333c-0.34666666666666757-0.3416666666666668-0.8633333333333333-0.6050000000000004-1.4699999999999989-0.6050000000000004-1.0550000000000033 8.881784197001252e-16-1.9050000000000011 0.745000000000001-1.9050000000000011 1.6666666666666679v5c-7.7666666666666675 0-10 8.116666666666669-10 14.166666666666666z m-5 7.5h23.333333333333336c0.9216666666666669 0 1.6666666666666643-0.7466666666666697 1.6666666666666643-1.6666666666666643v-10.076666666666668c-1.1066666666666691 1.1266666666666652-2.2733333333333334 2.3216666666666654-3.333333333333332 3.411666666666669v5h-20.000000000000004v-20.00166666666667h11.666666666666668v-3.333333333333334h-13.333333333333334c-0.9216666666666669 0-1.666666666666667 0.7466666666666661-1.666666666666667 1.666666666666666v23.333333333333336c0 0.9200000000000017 0.7450000000000001 1.6666666666666643 1.666666666666667 1.6666666666666643z' })
                )
            );
        }
    }]);

    return TiExport;
}(React.Component);

exports.default = TiExport;
module.exports = exports['default'];