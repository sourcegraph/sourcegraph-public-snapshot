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

var MdDirectionsWalk = function (_React$Component) {
    _inherits(MdDirectionsWalk, _React$Component);

    function MdDirectionsWalk() {
        _classCallCheck(this, MdDirectionsWalk);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdDirectionsWalk).apply(this, arguments));
    }

    _createClass(MdDirectionsWalk, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm16.328333333333337 14.843333333333335l-4.688333333333333 23.516666666666666h3.5166666666666657l3.044999999999998-13.361666666666665 3.4400000000000013 3.3583333333333343v10h3.3583333333333307v-12.5l-3.5166666666666657-3.356666666666669 1.0166666666666657-5q3.5933333333333337 4.138333333333335 9.14 4.138333333333335v-3.2833333333333314q-4.843333333333334 0-7.109999999999999-4.061666666666667l-1.7166666666666686-2.6566666666666663q-1.173333333333332-1.6400000000000006-2.8166666666666664-1.6400000000000006-0.23333333333333428 0-0.6616666666666653 0.07833333333333314t-0.6616666666666653 0.07833333333333314l-8.673333333333336 3.671666666666665v7.813333333333334h3.3599999999999994v-5.625l2.9666666666666686-1.1716666666666669z m6.171666666666663-5.703333333333335q-1.3283333333333331 0-2.3433333333333337-0.9766666666666666t-1.0166666666666657-2.3049999999999997 1.0166666666666657-2.3433333333333333 2.3433333333333337-1.015000000000001 2.3433333333333337 1.0166666666666666 1.0166666666666657 2.341666666666667-1.0166666666666657 2.3050000000000006-2.3433333333333337 0.9766666666666666z' })
                )
            );
        }
    }]);

    return MdDirectionsWalk;
}(React.Component);

exports.default = MdDirectionsWalk;
module.exports = exports['default'];