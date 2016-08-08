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

var MdPhotoSizeSelectLarge = function (_React$Component) {
    _inherits(MdPhotoSizeSelectLarge, _React$Component);

    function MdPhotoSizeSelectLarge() {
        _classCallCheck(this, MdPhotoSizeSelectLarge);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdPhotoSizeSelectLarge).apply(this, arguments));
    }

    _createClass(MdPhotoSizeSelectLarge, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm5 31.640000000000004h16.64l-5.313333333333333-7.109999999999999-4.140000000000001 5.390000000000001-3.046666666666667-3.5933333333333337z m-3.36-13.280000000000005h23.36v16.64h-20q-1.3283333333333331 0-2.3433333333333333-1.0166666666666657t-1.0166666666666666-2.3416666666666686v-13.283333333333331z m6.720000000000001-13.36h3.283333333333333v3.3599999999999994h-3.283333333333333v-3.3599999999999994z m6.639999999999999 0h3.3599999999999994v3.3599999999999994h-3.3599999999999994v-3.3599999999999994z m-10 0v3.3599999999999994h-3.36q0-1.25 1.0550000000000002-2.3049999999999997t2.3049999999999997-1.0549999999999997z m23.36 26.64h3.2833333333333314v3.3599999999999994h-3.2833333333333314v-3.3599999999999994z m0-26.64h3.2833333333333314v3.3599999999999994h-3.2833333333333314v-3.3599999999999994z m-26.72 6.640000000000001h3.3599999999999994v3.3599999999999994h-3.36v-3.3599999999999994z m33.36-6.640000000000001q1.25 0 2.3049999999999997 1.0549999999999997t1.0549999999999997 2.3049999999999997h-3.3599999999999994v-3.3599999999999994z m0 6.640000000000001h3.3599999999999994v3.3599999999999994h-3.3599999999999994v-3.3599999999999994z m-13.36-6.640000000000001h3.3599999999999994v3.3599999999999994h-3.3599999999999994v-3.3599999999999994z m16.72 26.64q0 1.25-1.0549999999999997 2.3049999999999997t-2.3049999999999997 1.0549999999999997v-3.3599999999999994h3.3599999999999994z m-3.3599999999999994-13.280000000000001h3.3599999999999994v3.2833333333333314h-3.3599999999999994v-3.2833333333333314z m0 6.640000000000001h3.3599999999999994v3.3599999999999994h-3.3599999999999994v-3.3599999999999994z' })
                )
            );
        }
    }]);

    return MdPhotoSizeSelectLarge;
}(React.Component);

exports.default = MdPhotoSizeSelectLarge;
module.exports = exports['default'];