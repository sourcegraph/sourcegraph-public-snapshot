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

var MdAlarmOff = function (_React$Component) {
    _inherits(MdAlarmOff, _React$Component);

    function MdAlarmOff() {
        _classCallCheck(this, MdAlarmOff);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdAlarmOff).apply(this, arguments));
    }

    _createClass(MdAlarmOff, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm13.360000000000001 5.466666666666667l-1.4066666666666663 1.1733333333333338-2.4200000000000017-2.3433333333333337 1.4833333333333325-1.1716666666666669z m14.061666666666666 25.15833333333333l-16.405-16.40833333333333q-2.6583333333333314 3.2833333333333314-2.6583333333333314 7.423333333333332 0 4.843333333333334 3.4000000000000004 8.283333333333331t8.24 3.4366666666666674q4.061666666666667 0 7.420000000000002-2.7333333333333343z m-22.578333333333333-26.796666666666667q5.9399999999999995 5.938333333333334 16.913333333333334 16.875000000000004t13.866666666666667 13.828333333333337l-2.1099999999999994 2.1083333333333343-3.671666666666667-3.671666666666667q-4.296666666666667 3.673333333333325-9.841666666666669 3.673333333333325-6.25 0-10.626666666666667-4.413333333333334t-4.373333333333333-10.586666666666666q0-5.466666666666669 3.67-9.766666666666667l-1.3283333333333331-1.3266666666666662-1.875 1.5616666666666674-2.3433333333333333-2.421666666666667 1.8766666666666665-1.4816666666666656-2.266666666666667-2.2666666666666675z m31.796666666666674 5.705l-2.1099999999999994 2.576666666666666-7.656666666666666-6.483333333333333 2.1099999999999994-2.5z m-16.640000000000008 0.4666666666666668q-2.1099999999999994 0-3.9833333333333343 0.7033333333333331l-2.58-2.5q3.283333333333333-1.5633333333333335 6.563333333333334-1.5633333333333335 6.25 0 10.625 4.413333333333335t4.375 10.586666666666666q0 3.4383333333333326-1.4833333333333343 6.563333333333333l-2.5799999999999983-2.5q0.7033333333333331-1.875 0.7033333333333331-4.063333333333333 0-4.843333333333334-3.3999999999999986-8.241666666666667t-8.240000000000002-3.3983333333333334z' })
                )
            );
        }
    }]);

    return MdAlarmOff;
}(React.Component);

exports.default = MdAlarmOff;
module.exports = exports['default'];