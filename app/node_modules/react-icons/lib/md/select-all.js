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

var MdSelectAll = function (_React$Component) {
    _inherits(MdSelectAll, _React$Component);

    function MdSelectAll() {
        _classCallCheck(this, MdSelectAll);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdSelectAll).apply(this, arguments));
    }

    _createClass(MdSelectAll, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm15 15v10h10v-10h-10z m-3.3599999999999994 13.36v-16.71666666666667h16.716666666666665v16.71666666666667h-16.714999999999996z m13.36-20v-3.3599999999999994h3.3599999999999994v3.3599999999999994h-3.3599999999999994z m0 26.64v-3.3599999999999994h3.3599999999999994v3.3599999999999994h-3.3599999999999994z m6.640000000000001-6.640000000000001v-3.3599999999999994h3.3599999999999994v3.3599999999999994h-3.3599999999999994z m0-13.360000000000001v-3.3599999999999977h3.3599999999999994v3.3599999999999994h-3.3599999999999994z m0 20v-3.3599999999999994h3.3599999999999994q0 1.3283333333333331-1.0166666666666657 2.3433333333333337t-2.3416666666666686 1.0166666666666657z m0-13.36v-3.2833333333333314h3.3599999999999994v3.2833333333333314h-3.3599999999999994z m-13.280000000000001 13.36v-3.3599999999999994h3.2833333333333314v3.3599999999999994h-3.2833333333333314z m-3.3599999999999994-30v3.3599999999999994h-3.3599999999999994v-3.3599999999999994h3.3599999999999994z m-10 23.36v-3.3599999999999994h3.3599999999999994v3.3599999999999994h-3.3599999999999994z m3.360000000000001 6.640000000000001q-1.3283333333333331 0-2.3433333333333337-1.0166666666666657t-1.0166666666666675-2.34h3.3616666666666664v3.3566666666666656z m23.28-30q1.3283333333333402 0 2.3433333333333337 1.0166666666666666t1.0166666666666657 2.3400000000000007h-3.361666666666668v-3.3566666666666674z m-9.999999999999996 0v3.3599999999999994h-3.283333333333335v-3.3599999999999994h3.2833333333333314z m-16.640000000000004 10v-3.3599999999999994h3.3599999999999994v3.3599999999999994h-3.3599999999999994z m6.640000000000001 20v-3.3599999999999994h3.3599999999999994v3.3599999999999994h-3.3599999999999994z m-6.640000000000001-13.36v-3.2833333333333314h3.3599999999999994v3.2833333333333314h-3.3599999999999994z m0-13.280000000000001q0-1.3283333333333331 1.0166666666666666-2.3433333333333337t2.3400000000000007-1.0166666666666657v3.3616666666666664h-3.3566666666666674z' })
                )
            );
        }
    }]);

    return MdSelectAll;
}(React.Component);

exports.default = MdSelectAll;
module.exports = exports['default'];