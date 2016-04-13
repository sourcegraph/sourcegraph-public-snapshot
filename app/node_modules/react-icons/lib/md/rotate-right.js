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

var MdRotateRight = function (_React$Component) {
    _inherits(MdRotateRight, _React$Component);

    function MdRotateRight() {
        _classCallCheck(this, MdRotateRight);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdRotateRight).apply(this, arguments));
    }

    _createClass(MdRotateRight, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm28.125 25.783333333333335q1.3283333333333331-1.8000000000000007 1.7166666666666686-4.141666666666666h3.3616666666666646q-0.46666666666666856 3.5933333333333337-2.6566666666666663 6.483333333333334z m-6.483333333333334 4.059999999999999q2.3416666666666686-0.39000000000000057 4.138333333333335-1.7166666666666686l2.421666666666667 2.4200000000000017q-2.8116666666666674 2.1883333333333326-6.561666666666667 2.6566666666666663v-3.3599999999999994z m11.561666666666667-11.483333333333334h-3.3599999999999994q-0.39000000000000057-2.3433333333333337-1.7166666666666686-4.140000000000001l2.4200000000000017-2.3433333333333337q2.1883333333333326 2.8900000000000006 2.6566666666666663 6.483333333333334z m-7.266666666666666-9.141666666666667l-7.576666666666668 7.423333333333334v-6.483333333333334q-3.5933333333333337 0.620000000000001-5.976666666666667 3.391666666666671t-2.383333333333333 6.4499999999999975 2.383333333333333 6.443333333333335 5.976666666666667 3.3999999999999986v3.3583333333333343q-4.921666666666667-0.6233333333333348-8.32-4.373333333333335t-3.3999999999999986-8.828333333333333 3.4016666666666673-8.833333333333332 8.32-4.375v-5.153333333333334z' })
                )
            );
        }
    }]);

    return MdRotateRight;
}(React.Component);

exports.default = MdRotateRight;
module.exports = exports['default'];