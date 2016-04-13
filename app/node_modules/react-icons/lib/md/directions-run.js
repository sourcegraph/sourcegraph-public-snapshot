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

var MdDirectionsRun = function (_React$Component) {
    _inherits(MdDirectionsRun, _React$Component);

    function MdDirectionsRun() {
        _classCallCheck(this, MdDirectionsRun);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdDirectionsRun).apply(this, arguments));
    }

    _createClass(MdDirectionsRun, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm16.483333333333334 32.266666666666666l-11.636666666666667-2.2666666666666657 0.6249999999999991-3.3599999999999994 8.203333333333333 1.6400000000000006 2.6566666666666663-13.516666666666666-3.046666666666667 1.1733333333333338v5.703333333333335h-3.285v-7.813333333333338l8.673333333333332-3.671666666666667q0.23333333333333428 0 0.663333333333334-0.07833333333333314t0.663333333333334-0.07666666666666622q1.6400000000000006 0 2.8133333333333326 1.6383333333333336l1.6400000000000006 2.6566666666666663q2.2666666666666657 3.9833333333333325 7.188333333333333 3.9833333333333325v3.361666666666668q-5.546666666666667 0-9.14-4.140000000000001l-1.0133333333333319 5 3.5166666666666657 3.2833333333333314v12.5h-3.361666666666668v-10l-3.518333333333331-3.2833333333333314z m6.016666666666666-23.12833333333333q-1.3283333333333331-1.7763568394002505e-15-2.3433333333333337-1.0166666666666693t-1.0166666666666657-2.341666666666667 1.0166666666666657-2.3066666666666666 2.3433333333333337-0.9733333333333327 2.3049999999999997 0.9766666666666666 0.9766666666666666 2.3049999999999997-0.9783333333333317 2.341666666666667-2.3049999999999997 1.0166666666666675z' })
                )
            );
        }
    }]);

    return MdDirectionsRun;
}(React.Component);

exports.default = MdDirectionsRun;
module.exports = exports['default'];