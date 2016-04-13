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

var MdSimCard = function (_React$Component) {
    _inherits(MdSimCard, _React$Component);

    function MdSimCard() {
        _classCallCheck(this, MdSimCard);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdSimCard).apply(this, arguments));
    }

    _createClass(MdSimCard, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm28.36 25v-6.640000000000001h-3.3599999999999994v6.640000000000001h3.3599999999999994z m-6.719999999999999-3.3599999999999994v-3.2833333333333314h-3.2833333333333314v3.2833333333333314h3.2833333333333314z m0 10v-6.640000000000001h-3.2833333333333314v6.640000000000001h3.2833333333333314z m-6.640000000000001-6.640000000000001v-6.640000000000001h-3.3599999999999994v6.640000000000001h3.3599999999999994z m13.36 6.640000000000001v-3.2833333333333314h-3.3599999999999994v3.2833333333333314h3.3599999999999994z m-13.360000000000001 0v-3.2833333333333314h-3.3599999999999977v3.2833333333333314h3.3599999999999994z m18.28333333333333-25l0.07666666666666799 26.71666666666667q0 1.3299999999999983-1.0166666666666657 2.306666666666665t-2.3416666666666686 0.9766666666666666h-20.001666666666665q-1.3266666666666662 0-2.341666666666667-0.9766666666666666t-1.0166666666666666-2.3049999999999997v-20l10-10h13.361666666666665q1.3283333333333331 0 2.3049999999999997 0.9766666666666666t0.9766666666666666 2.3066666666666666z' })
                )
            );
        }
    }]);

    return MdSimCard;
}(React.Component);

exports.default = MdSimCard;
module.exports = exports['default'];