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

var MdMemory = function (_React$Component) {
    _inherits(MdMemory, _React$Component);

    function MdMemory() {
        _classCallCheck(this, MdMemory);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdMemory).apply(this, arguments));
    }

    _createClass(MdMemory, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm28.36 28.36v-16.71666666666667h-16.71666666666667v16.71666666666667h16.71666666666667z m6.640000000000001-10h-3.3599999999999994v3.2833333333333314h3.3599999999999994v3.356666666666669h-3.3599999999999994v3.361666666666668q0 1.3283333333333331-0.9766666666666666 2.3049999999999997t-2.3049999999999997 0.9766666666666666h-3.3583333333333343v3.360000000000003h-3.361666666666668v-3.3599999999999994h-3.283333333333335v3.3599999999999994h-3.354999999999997v-3.3599999999999994h-3.3633333333333333q-1.3283333333333331 0-2.3049999999999997-0.9766666666666666t-0.9783333333333335-2.3049999999999997v-3.3616666666666717h-3.3583333333333334v-3.3583333333333343h3.3600000000000003v-3.2833333333333314h-3.3600000000000003v-3.3583333333333343h3.3600000000000003v-3.3599999999999994q0-1.3283333333333331 0.9766666666666666-2.3049999999999997t2.303333333333333-0.9749999999999996h3.360000000000001v-3.360000000000001h3.3599999999999994v3.3583333333333343h3.283333333333335v-3.3583333333333343h3.3583333333333343v3.3583333333333343h3.3583333333333343q1.3283333333333331 0 2.3049999999999997 0.9766666666666666t0.9766666666666666 2.3066666666666666v3.3583333333333325h3.3633333333333297v3.3599999999999994z m-13.36 3.280000000000001v-3.2833333333333314h-3.2833333333333314v3.2833333333333314h3.2833333333333314z m3.3599999999999994-6.640000000000001v10h-10v-10h10z' })
                )
            );
        }
    }]);

    return MdMemory;
}(React.Component);

exports.default = MdMemory;
module.exports = exports['default'];