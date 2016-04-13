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

var GoOctoface = function (_React$Component) {
    _inherits(GoOctoface, _React$Component);

    function GoOctoface() {
        _classCallCheck(this, GoOctoface);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(GoOctoface).apply(this, arguments));
    }

    _createClass(GoOctoface, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm36.75 10.8475c0.322499999999998-0.7899999999999991 1.3812500000000014-3.9749999999999996-0.33500000000000085-8.2775 0 0-2.6312500000000014-0.8325-8.587499999999999 3.2375-2.49625-0.6849999999999996-5.177500000000002-0.7874999999999996-7.828749999999999-0.7874999999999996s-5.331250000000001 0.10250000000000004-7.83375 0.7912499999999998c-5.955000000000001-4.07375-8.58375-3.2425-8.58375-3.2425-1.7150000000000005 4.30625-0.6562500000000009 7.48875-0.3337500000000011 8.27875-2.01875 2.182500000000001-3.25 4.96875-3.25 8.385000000000002 3.907464629637758e-16 12.86 8.330000000000002 15.767499999999998 19.955000000000002 15.767499999999998 11.630000000000003 0 20.045-2.907499999999999 20.045-15.767500000000002 0-3.41625-1.2299999999999969-6.202500000000001-3.25-8.385z m-16.75 21.69c-8.2575 0-14.9525-0.384999999999998-14.9525-8.384999999999998 0-1.9125000000000014 0.9399999999999995-3.6950000000000003 2.5525-5.168749999999999 2.6850000000000005-2.460000000000001 7.237499999999999-1.1600000000000001 12.4-1.1600000000000001 5.166249999999998 0 9.712499999999999-1.3000000000000007 12.399999999999999 1.15625 1.6124999999999972 1.4750000000000014 2.5562499999999986 3.254999999999999 2.5562499999999986 5.168749999999999 0 8.002500000000001-6.699999999999999 8.3875-14.95625 8.3875z m-6.282499999999999-12.52c-1.6587499999999995 0-3.00375 1.995000000000001-3.00375 4.460000000000001s1.3462499999999995 4.465 3.004999999999999 4.465c1.6549999999999994 0 3-2 3-4.465s-1.3450000000000006-4.460000000000001-3-4.460000000000001z m12.56625 0c-1.6550000000000011 0-3.0025000000000013 1.995000000000001-3.0025000000000013 4.460000000000001s1.3475000000000001 4.465 3.0025000000000013 4.465 3-2 3-4.465c0.002500000000001279-2.4662499999999987-1.3425000000000011-4.460000000000001-3-4.460000000000001z' })
                )
            );
        }
    }]);

    return GoOctoface;
}(React.Component);

exports.default = GoOctoface;
module.exports = exports['default'];