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

var MdRefresh = function (_React$Component) {
    _inherits(MdRefresh, _React$Component);

    function MdRefresh() {
        _classCallCheck(this, MdRefresh);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdRefresh).apply(this, arguments));
    }

    _createClass(MdRefresh, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm29.453333333333337 10.546666666666667l3.90666666666667-3.9066666666666663v11.716666666666669h-11.716666666666669l5.388333333333335-5.388333333333334q-2.970000000000006-2.9683333333333355-7.031666666666673-2.9683333333333355-4.139999999999999 0-7.07 2.9299999999999997t-2.9299999999999997 7.07 2.9299999999999997 7.07 7.07 2.9299999999999997q3.2833333333333314 0 5.859999999999999-1.836666666666666t3.5933333333333337-4.805h3.4383333333333326q-1.0933333333333337 4.375-4.688333333333333 7.188333333333333t-8.203333333333333 2.8133333333333326q-5.466666666666667 0-9.375-3.9066666666666663t-3.9083333333333323-9.453333333333333 3.908333333333334-9.453333333333333 9.374999999999998-3.9066666666666663q5.546666666666667 0 9.453333333333333 3.9066666666666663z' })
                )
            );
        }
    }]);

    return MdRefresh;
}(React.Component);

exports.default = MdRefresh;
module.exports = exports['default'];