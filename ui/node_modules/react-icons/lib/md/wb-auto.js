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

var MdWbAuto = function (_React$Component) {
    _inherits(MdWbAuto, _React$Component);

    function MdWbAuto() {
        _classCallCheck(this, MdWbAuto);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdWbAuto).apply(this, arguments));
    }

    _createClass(MdWbAuto, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm17.188333333333336 26.64h3.125l-5.313333333333336-15h-3.3599999999999994l-5.3133333333333335 15h3.203333333333334l1.1716666666666669-3.2833333333333314h5.316666666666666z m19.45-15h3.049999999999997l-3.4399999999999977 15h-2.8916666666666657l-2.5-10.156666666666666-2.5 10.156666666666666h-2.9666666666666686l-0.158333333333335-0.7033333333333331q-1.6400000000000006 3.3599999999999994-4.843333333333334 5.390000000000001t-7.033333333333333 2.0333333333333314q-5.545 0-9.450000000000001-3.9466666666666654t-3.908333333333331-9.413333333333334 3.9050000000000002-9.413333333333334 9.453333333333333-3.9450000000000003q6.406666666666666 0 10.39 5h1.254999999999999l2.030000000000001 10.546666666666667 2.5-10.546666666666667h2.6566666666666663l2.5 10.546666666666667z m-25.233333333333334 9.453333333333333l1.9566666666666652-6.093333333333334 1.875 6.093333333333334h-3.828333333333333z' })
                )
            );
        }
    }]);

    return MdWbAuto;
}(React.Component);

exports.default = MdWbAuto;
module.exports = exports['default'];