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

var MdImportContacts = function (_React$Component) {
    _inherits(MdImportContacts, _React$Component);

    function MdImportContacts() {
        _classCallCheck(this, MdImportContacts);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdImportContacts).apply(this, arguments));
    }

    _createClass(MdImportContacts, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm35 30.86v-19.21666666666667q-2.578333333333333-0.7833333333333314-5.859999999999999-0.7833333333333314-5.078333333333333 0-9.14 2.5v19.14q4.063333333333333-2.5 9.14-2.5 3.046666666666667 0 5.859999999999999 0.8599999999999994z m-5.859999999999996-23.36q5.938333333333336 0 9.216666666666665 2.5v24.296666666666667q0 0.3133333333333326-0.2716666666666683 0.586666666666666t-0.5833333333333357 0.27333333333333343q-0.23666666666666458 0-0.3916666666666657-0.07833333333333314-3.203333333333333-1.7166666666666686-7.966666666666669-1.7166666666666686-5.079999999999998 0-9.141666666666666 2.5-3.3599999999999994-2.5-9.14-2.5-4.216666666666667 0-7.966666666666668 1.7950000000000017-0.08000000000000007 0-0.19666666666666677 0.038333333333333997t-0.19499999999999984 0.038333333333333997q-0.31166666666666654 0-0.5833333333333335-0.23333333333333428t-0.2749999999999999-0.5499999999999972v-24.450000000000003q3.355000000000003-2.5 9.213333333333335-2.5 5.785000000000004 0 9.14166666666667 2.5 3.3599999999999994-2.5 9.14-2.5z' })
                )
            );
        }
    }]);

    return MdImportContacts;
}(React.Component);

exports.default = MdImportContacts;
module.exports = exports['default'];