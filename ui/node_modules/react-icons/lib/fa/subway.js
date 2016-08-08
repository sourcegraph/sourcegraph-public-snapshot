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

var FaSubway = function (_React$Component) {
    _inherits(FaSubway, _React$Component);

    function FaSubway() {
        _classCallCheck(this, FaSubway);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaSubway).apply(this, arguments));
    }

    _createClass(FaSubway, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm27.142857142857142 0q4.12857142857143 0 7.064285714285713 2.085714285714286t2.9357142857142904 5.057142857142857v20q0 2.8999999999999986-2.799999999999997 4.957142857142859t-6.82 2.164285714285711l4.7542857142857144 4.508571428571429q0.3571428571428541 0.33571428571428896 0.1785714285714306 0.7828571428571465t-0.6714285714285708 0.4471428571428575h-23.570000000000007q-0.4914285714285711 0-0.6714285714285717-0.4457142857142884t0.17999999999999972-0.7814285714285703l4.754285714285715-4.50714285714286q-4.017142857142858-0.11142857142856855-6.8185714285714285-2.1657142857142873t-2.801428571428573-4.959999999999997v-20q0-2.968571428571428 2.9357142857142855-5.057142857142856t7.064285714285715-2.0857142857142863h14.285714285714288z m-17.857142857142854 29.285714285714285q1.4714285714285698 0 2.5228571428571414-1.048571428571428t1.048571428571428-2.5228571428571414-1.048571428571428-2.5228571428571414-2.522857142857143-1.0485714285714316-2.522857142857143 1.048571428571428-1.048571428571429 2.522857142857145 1.048571428571429 2.5228571428571414 2.522857142857143 1.048571428571428z m9.285714285714285-12.142857142857142v-11.428571428571427h-12.142857142857144v11.428571428571427h12.142857142857144z m12.142857142857142 12.142857142857142q1.471428571428568 0 2.5228571428571414-1.048571428571428t1.048571428571428-2.5228571428571414-1.048571428571428-2.5228571428571414-2.5228571428571414-1.0485714285714316-2.5228571428571414 1.048571428571428-1.0485714285714316 2.522857142857145 1.048571428571428 2.5228571428571414 2.522857142857145 1.048571428571428z m3.5714285714285694-12.142857142857142v-11.428571428571427h-12.857142857142858v11.428571428571427h12.857142857142858z' })
                )
            );
        }
    }]);

    return FaSubway;
}(React.Component);

exports.default = FaSubway;
module.exports = exports['default'];