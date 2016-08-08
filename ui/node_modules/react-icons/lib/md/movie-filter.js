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

var MdMovieFilter = function (_React$Component) {
    _inherits(MdMovieFilter, _React$Component);

    function MdMovieFilter() {
        _classCallCheck(this, MdMovieFilter);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdMovieFilter).apply(this, arguments));
    }

    _createClass(MdMovieFilter, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm25.156666666666666 20.313333333333336l1.9533333333333331-0.8599999999999994-1.9533333333333331-0.8599999999999994-0.8599999999999994-1.9533333333333331-0.8599999999999994 1.9533333333333331-1.9533333333333331 0.8599999999999994 1.9533333333333331 0.8599999999999994 0.8599999999999994 1.9533333333333331z m-6.093333333333334 5.858333333333334l3.828333333333333-1.7966666666666669-3.828333333333333-1.716666666666665-1.7966666666666669-3.908333333333335-1.7166666666666668 3.9066666666666663-3.908333333333333 1.716666666666665 3.9066666666666663 1.8000000000000007 1.7183333333333337 3.8266666666666644z m10.936666666666667-19.53333333333334h6.640000000000001v23.361666666666668q0 1.3283333333333331-0.9766666666666666 2.3433333333333337t-2.3049999999999997 1.0166666666666657h-26.71666666666667q-1.3299999999999992 0-2.3066666666666658-1.0166666666666657t-0.9749999999999996-2.3433333333333337v-20q0-1.3283333333333331 0.9766666666666666-2.3433333333333337t2.3066666666666666-1.0166666666666666h1.716666666666666l3.283333333333333 6.72h4.999999999999998l-3.283333333333333-6.716666666666668h3.283333333333333l3.356666666666669 6.7150000000000025h5l-3.3599999999999994-6.716666666666668h3.3599999999999994l3.3599999999999994 6.716666666666668h5z' })
                )
            );
        }
    }]);

    return MdMovieFilter;
}(React.Component);

exports.default = MdMovieFilter;
module.exports = exports['default'];