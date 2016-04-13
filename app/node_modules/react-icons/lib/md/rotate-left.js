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

var MdRotateLeft = function (_React$Component) {
    _inherits(MdRotateLeft, _React$Component);

    function MdRotateLeft() {
        _classCallCheck(this, MdRotateLeft);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdRotateLeft).apply(this, arguments));
    }

    _createClass(MdRotateLeft, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm21.64 6.796666666666668q4.921666666666667 0.625 8.32 4.374999999999999t3.3999999999999986 8.828333333333333-3.3999999999999986 8.828333333333333-8.32 4.375v-3.3599999999999994q3.5933333333333337-0.625 5.976666666666667-3.3999999999999986t2.383333333333333-6.443333333333335-2.383333333333333-6.445-5.976666666666667-3.4000000000000004v6.486666666666666l-7.578333333333335-7.421666666666665 7.578333333333335-7.578333333333334v5.158333333333334z m-9.843333333333334 23.75l2.421666666666667-2.421666666666667q1.8000000000000007 1.3283333333333331 4.141666666666666 1.7166666666666686v3.3616666666666646q-3.75-0.46666666666666856-6.563333333333333-2.6566666666666663z m-1.6400000000000006-8.906666666666666q0.39000000000000057 2.2666666666666657 1.6400000000000006 4.140000000000001l-2.3433333333333337 2.3433333333333337q-2.1883333333333335-2.8900000000000006-2.6566666666666663-6.483333333333334h3.3599999999999994z m1.7166666666666668-7.421666666666667q-1.4049999999999994 2.0333333333333314-1.7166666666666668 4.141666666666666h-3.3599999999999985q0.4666666666666668-3.5166666666666657 2.7333333333333334-6.483333333333334z' })
                )
            );
        }
    }]);

    return MdRotateLeft;
}(React.Component);

exports.default = MdRotateLeft;
module.exports = exports['default'];