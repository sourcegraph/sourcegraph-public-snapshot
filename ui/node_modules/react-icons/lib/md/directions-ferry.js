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

var MdDirectionsFerry = function (_React$Component) {
    _inherits(MdDirectionsFerry, _React$Component);

    function MdDirectionsFerry() {
        _classCallCheck(this, MdDirectionsFerry);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdDirectionsFerry).apply(this, arguments));
    }

    _createClass(MdDirectionsFerry, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm10 10v6.640000000000001l10-3.283333333333333 10 3.283333333333333v-6.640000000000001h-20z m-3.4366666666666665 21.640000000000004l-3.125-11.093333333333334q-0.5466666666666669-1.5633333333333326 1.0933333333333337-2.1099999999999994l2.1083333333333334-0.7033333333333331v-7.733333333333338q0-1.33 1.0166666666666666-2.3450000000000006t2.343333333333333-1.0166666666666657h5v-5h10v5h5q1.3283333333333331 0 2.3433333333333337 1.0166666666666666t1.0166666666666657 2.3449999999999998v7.733333333333334l2.1083333333333343 0.7033333333333331q1.6416666666666657 0.5466666666666669 1.0949999999999989 2.1099999999999994l-3.125 11.093333333333334h-0.07833333333333314q-3.828333333333333 0-6.716666666666669-3.2833333333333314-2.8916666666666657 3.2833333333333314-6.641666666666666 3.2833333333333314t-6.640000000000001-3.2833333333333314q-2.8900000000000006 3.2833333333333314-6.716666666666668 3.2833333333333314h-0.08000000000000007z m26.796666666666667 3.359999999999996h3.2833333333333314v3.3599999999999994h-3.2833333333333314q-3.5166666666666657 0-6.716666666666669-1.6400000000000006-6.641666666666666 3.4383333333333326-13.283333333333333 0-3.203333333333333 1.6400000000000006-6.716666666666668 1.6400000000000006h-3.283333333333333v-3.3599999999999994h3.283333333333333q3.591666666666666 0 6.716666666666668-2.1883333333333326 3.0466666666666686 2.1099999999999994 6.640000000000002 2.1099999999999994t6.640000000000001-2.1099999999999994q3.126666666666665 2.1883333333333326 6.716666666666669 2.1883333333333326z' })
                )
            );
        }
    }]);

    return MdDirectionsFerry;
}(React.Component);

exports.default = MdDirectionsFerry;
module.exports = exports['default'];