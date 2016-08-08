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

var MdCloudOff = function (_React$Component) {
    _inherits(MdCloudOff, _React$Component);

    function MdCloudOff() {
        _classCallCheck(this, MdCloudOff);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdCloudOff).apply(this, arguments));
    }

    _createClass(MdCloudOff, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm12.89 16.64h-2.8900000000000006q-2.7333333333333334 0-4.6883333333333335 1.9916666666666671t-1.9533333333333331 4.725000000000001 1.9533333333333331 4.688333333333333 4.6883333333333335 1.9549999999999983h16.25z m-7.890000000000001-7.890000000000001l2.1100000000000003-2.1100000000000003 27.89 27.89-2.1099999999999994 2.1099999999999994-3.3599999999999994-3.2833333333333314h-19.53333333333333q-4.138333333333335 0-7.066666666666669-2.9283333333333346t-2.9333333333333336-7.07q0-4.063333333333333 2.816666666666667-6.953333333333333t6.795-3.046666666666667z m27.266666666666666 7.966666666666669q3.200000000000003 0.23666666666666814 5.466666666666669 2.616666666666667t2.2666666666666657 5.666666666666664q0 4.295000000000002-3.5166666666666657 6.795000000000002l-2.421666666666667-2.421666666666667q2.578333333333333-1.4066666666666663 2.578333333333333-4.375 0-2.0333333333333314-1.4833333333333343-3.5166666666666657t-3.5166666666666657-1.4833333333333343h-2.5v-0.8616666666666681q0-3.828333333333333-2.6566666666666663-6.483333333333334t-6.483333333333334-2.6583333333333314q-2.423333333333332 0-4.220000000000001 1.0166666666666657l-2.5-2.423333333333334q3.0466666666666686-1.9533333333333331 6.716666666666667-1.9533333333333331 4.533333333333335 0 7.970000000000002 2.8499999999999996t4.296666666666667 7.228333333333332z' })
                )
            );
        }
    }]);

    return MdCloudOff;
}(React.Component);

exports.default = MdCloudOff;
module.exports = exports['default'];