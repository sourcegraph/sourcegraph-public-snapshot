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

var MdDeveloperBoard = function (_React$Component) {
    _inherits(MdDeveloperBoard, _React$Component);

    function MdDeveloperBoard() {
        _classCallCheck(this, MdDeveloperBoard);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdDeveloperBoard).apply(this, arguments));
    }

    _createClass(MdDeveloperBoard, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm20 18.36h6.640000000000001v10h-6.640000000000001v-10z m-10-6.719999999999999h8.36v8.36h-8.36v-8.36z m10 0h6.640000000000001v5h-6.640000000000001v-5z m-10 10h8.36v6.716666666666669h-8.36v-6.716666666666669z m20 10v-23.28333333333333h-23.36v23.28333333333333h23.36z m6.640000000000008-16.64h-3.2833333333333314v3.3599999999999994h3.2833333333333314v3.2833333333333314h-3.2833333333333314v3.356666666666669h3.2833333333333314v3.361666666666668h-3.2833333333333314v3.283333333333335q0 1.326666666666668-1.0133333333333354 2.3416666666666686t-2.3433333333333337 1.0166666666666657h-23.360000000000007q-1.3283333333333331 0-2.3049999999999997-1.0166666666666657t-0.9750000000000001-2.3433333333333337v-23.28333333333334q0-1.3266666666666653 0.976666666666667-2.341666666666665t2.3066666666666666-1.0133333333333336h23.356666666666666q1.3283333333333331 0 2.3433333333333337 1.0166666666666666t1.0166666666666657 2.341666666666666v3.283333333333333h3.280000000000001v3.3533333333333335z' })
                )
            );
        }
    }]);

    return MdDeveloperBoard;
}(React.Component);

exports.default = MdDeveloperBoard;
module.exports = exports['default'];