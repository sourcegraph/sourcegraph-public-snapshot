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

var FaShield = function (_React$Component) {
    _inherits(FaShield, _React$Component);

    function FaShield() {
        _classCallCheck(this, FaShield);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaShield).apply(this, arguments));
    }

    _createClass(FaShield, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm30 21.42857142857143v-14.285714285714288h-10v25.38q2.6571428571428584-1.4057142857142857 4.7542857142857144-3.057142857142857 5.2457142857142856-4.108571428571427 5.2457142857142856-8.037142857142854z m4.285714285714285-17.142857142857142v17.142857142857142q0 1.9200000000000017-0.7471428571428547 3.805714285714288t-1.8528571428571432 3.3485714285714288-2.634285714285717 2.8457142857142834-2.8242857142857147 2.299999999999997-2.7028571428571446 1.7285714285714278-1.9971428571428582 1.1042857142857159-0.9485714285714302 0.4471428571428575q-0.26428571428570535 0.134285714285717-0.5785714285714221 0.134285714285717t-0.5799999999999983-0.134285714285717q-0.35714285714285765-0.15714285714285836-0.9485714285714302-0.4471428571428575t-1.9971428571428582-1.1042857142857159-2.6999999999999993-1.7285714285714278-2.8271428571428565-2.299999999999997-2.6328571428571426-2.845714285714287-1.852857142857144-3.348571428571425-0.7471428571428564-3.805714285714288v-17.142857142857146q0-0.5799999999999992 0.4242857142857144-1.0042857142857136t1.0042857142857144-0.42428571428571393h25.71428571428572q0.5799999999999983 0 1.0042857142857144 0.4242857142857144t0.42428571428570905 1.004285714285714z' })
                )
            );
        }
    }]);

    return FaShield;
}(React.Component);

exports.default = FaShield;
module.exports = exports['default'];