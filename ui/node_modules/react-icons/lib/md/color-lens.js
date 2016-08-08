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

var MdColorLens = function (_React$Component) {
    _inherits(MdColorLens, _React$Component);

    function MdColorLens() {
        _classCallCheck(this, MdColorLens);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdColorLens).apply(this, arguments));
    }

    _createClass(MdColorLens, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm29.140000000000004 20q1.0166666666666657 0 1.7583333333333329-0.7033333333333331t0.7399999999999984-1.7966666666666669-0.7416666666666671-1.7966666666666669-1.7583333333333329-0.7033333333333331-1.7583333333333329 0.7033333333333331-0.7433333333333323 1.7966666666666669 0.7416666666666671 1.7966666666666669 1.7566666666666677 0.7033333333333331z m-5-6.640000000000001q1.0166666666666657 0 1.7583333333333329-0.7416666666666671t0.7399999999999984-1.7566666666666677-0.7416666666666671-1.7583333333333329-1.7600000000000016-0.7400000000000002-1.7583333333333329 0.7416666666666671-0.7433333333333323 1.7599999999999998 0.7416666666666671 1.7583333333333329 1.7566666666666677 0.7433333333333341z m-8.280000000000001 0q1.0166666666666657 0 1.7583333333333329-0.7416666666666671t0.7433333333333323-1.7566666666666677-0.7416666666666671-1.7583333333333329-1.7566666666666677-0.7400000000000002-1.7583333333333329 0.7416666666666671-0.7400000000000002 1.7599999999999998 0.7416666666666671 1.7583333333333329 1.7599999999999998 0.7433333333333341z m-5 6.640000000000001q1.0166666666666657 0 1.7583333333333329-0.7033333333333331t0.7433333333333341-1.7966666666666669-0.7416666666666671-1.7966666666666669-1.7583333333333346-0.7033333333333331-1.7583333333333329 0.7033333333333331-0.7400000000000002 1.7966666666666669 0.7416666666666671 1.7966666666666669 1.756666666666666 0.7033333333333331z m9.139999999999997-15q6.171666666666667 0 10.586666666666666 3.9066666666666663t4.413333333333334 9.453333333333333q0 3.4383333333333326-2.461666666666666 5.859999999999999t-5.899999999999999 2.421666666666667h-2.8883333333333354q-1.0933333333333337 0-1.7966666666666669 0.7416666666666671t-0.7033333333333331 1.7583333333333329q0 0.8599999999999994 0.625 1.6400000000000006t0.625 1.7166666666666686q0 1.0933333333333337-0.7033333333333331 1.7966666666666669t-1.7966666666666669 0.7049999999999983q-6.25 0-10.625-4.375t-4.375-10.625 4.375-10.625 10.625-4.375z' })
                )
            );
        }
    }]);

    return MdColorLens;
}(React.Component);

exports.default = MdColorLens;
module.exports = exports['default'];