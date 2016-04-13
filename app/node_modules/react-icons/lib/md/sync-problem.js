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

var MdSyncProblem = function (_React$Component) {
    _inherits(MdSyncProblem, _React$Component);

    function MdSyncProblem() {
        _classCallCheck(this, MdSyncProblem);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdSyncProblem).apply(this, arguments));
    }

    _createClass(MdSyncProblem, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm18.36 21.64v-10h3.2833333333333314v10h-3.2833333333333314z m16.64-15l-3.9066666666666663 3.9833333333333343q3.9066666666666663 3.91 3.9066666666666663 9.376666666666665 0 4.609999999999999-2.8133333333333326 8.203333333333333t-7.186666666666667 4.688333333333333v-3.4383333333333326q2.9666666666666686-1.0166666666666657 4.803333333333335-3.5933333333333337t1.8366666666666696-5.859999999999999q0-4.140000000000001-2.8900000000000006-7.033333333333333l-3.7500000000000036 3.674999999999999v-10h10z m-16.64 21.720000000000002v-3.360000000000003h3.2833333333333314v3.3599999999999994h-3.2833333333333314z m-13.36-8.360000000000003q0-4.609999999999999 2.8133333333333335-8.203333333333333t7.1866666666666665-4.6883333333333335v3.4383333333333335q-2.966666666666667 1.0166666666666657-4.803333333333335 3.5933333333333337t-1.8366666666666642 5.859999999999999q0 4.140000000000001 2.8900000000000006 7.033333333333331l3.7499999999999982-3.674999999999997v10h-10l3.9066666666666663-3.9833333333333343q-3.9066666666666663-3.908333333333335-3.9066666666666663-9.375z' })
                )
            );
        }
    }]);

    return MdSyncProblem;
}(React.Component);

exports.default = MdSyncProblem;
module.exports = exports['default'];