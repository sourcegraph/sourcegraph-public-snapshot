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

var FaMusic = function (_React$Component) {
    _inherits(FaMusic, _React$Component);

    function FaMusic() {
        _classCallCheck(this, FaMusic);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(FaMusic).apply(this, arguments));
    }

    _createClass(FaMusic, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm37.142857142857146 5v25q0 1.1142857142857139-0.7571428571428598 1.985714285714284t-1.9214285714285708 1.351428571428574-2.3100000000000023 0.7142857142857153-2.154285714285713 0.23428571428571132-2.154285714285713-0.23428571428571132-2.3099999999999987-0.7142857142857153-1.9200000000000017-1.3500000000000014-0.7585714285714289-1.9871428571428567 0.7571428571428562-1.985714285714284 1.9214285714285708-1.3514285714285705 2.3099999999999987-0.7142857142857153 2.1542857142857166-0.23428571428571487q2.3428571428571416 0 4.285714285714285 0.8714285714285701v-11.988571428571428l-17.142857142857142 5.288571428571428v15.82857142857143q0 1.1142857142857139-0.7571428571428562 1.9857142857142875t-1.9214285714285708 1.3500000000000014-2.3100000000000005 0.7142857142857153-2.154285714285715 0.23571428571428044-2.154285714285715-0.23571428571428754-2.3100000000000005-0.7142857142857153-1.92142857142857-1.3499999999999943-0.7571428571428576-1.9857142857142875 0.7571428571428571-1.9885714285714258 1.9214285714285713-1.3500000000000014 2.3099999999999996-0.7142857142857153 2.154285714285715-0.2328571428571422q2.3428571428571434 0 4.285714285714285 0.8685714285714283v-21.58571428571429q0-0.6899999999999977 0.4242857142857144-1.259999999999998t1.0942857142857143-0.7928571428571427l18.571428571428577-5.714285714285714q0.2671428571428578-0.08714285714285763 0.6242857142857119-0.08714285714285763 0.894285714285715 0 1.518571428571427 0.6257142857142859t0.624285714285719 1.5142857142857142z' })
                )
            );
        }
    }]);

    return FaMusic;
}(React.Component);

exports.default = FaMusic;
module.exports = exports['default'];