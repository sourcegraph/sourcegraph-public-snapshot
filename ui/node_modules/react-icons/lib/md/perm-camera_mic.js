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

var MdPermCameraMic = function (_React$Component) {
    _inherits(MdPermCameraMic, _React$Component);

    function MdPermCameraMic() {
        _classCallCheck(this, MdPermCameraMic);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdPermCameraMic).apply(this, arguments));
    }

    _createClass(MdPermCameraMic, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm23.36 21.64v-6.640000000000001q0-1.3283333333333331-1.0166666666666657-2.3433333333333337t-2.3416666666666686-1.0166666666666657-2.3416666666666686 1.0166666666666657-1.0183333333333309 2.3433333333333337v6.640000000000001q0 1.3283333333333331 1.0166666666666657 2.3433333333333337t2.3416666666666686 1.0166666666666657 2.344999999999999-1.0166666666666657 1.0166666666666657-2.3433333333333337z m10-13.280000000000001q1.3283333333333331 0 2.3049999999999997 0.9766666666666666t0.9750000000000014 2.3049999999999997v20q0 1.3283333333333331-0.9766666666666666 2.3433333333333337t-2.306666666666665 1.0166666666666657h-11.716666666666669v-3.5166666666666657q3.5133333333333354-0.6233333333333348 5.936666666666667-3.3966666666666683t2.423333333333332-6.446666666666665h-3.3583333333333343q0 2.7333333333333343-1.9533333333333331 4.726666666666667t-4.688333333333333 1.9933333333333358-4.688333333333334-1.995000000000001-1.9533333333333331-4.725000000000001h-3.3583333333333325q0 3.671666666666667 2.42 6.445t5.938333333333334 3.3999999999999986v3.5133333333333354h-11.716666666666667q-1.33 0-2.3066666666666666-1.0166666666666657t-0.9750000000000001-2.3416666666666686v-20q0-1.3283333333333331 0.976666666666667-2.3049999999999997t2.3066666666666666-0.9766666666666666h5.313333333333333l3.043333333333333-3.3599999999999994h10l3.046666666666667 3.3599999999999994h5.313333333333333z' })
                )
            );
        }
    }]);

    return MdPermCameraMic;
}(React.Component);

exports.default = MdPermCameraMic;
module.exports = exports['default'];