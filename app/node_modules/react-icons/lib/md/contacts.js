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

var MdContacts = function (_React$Component) {
    _inherits(MdContacts, _React$Component);

    function MdContacts() {
        _classCallCheck(this, MdContacts);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdContacts).apply(this, arguments));
    }

    _createClass(MdContacts, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm28.36 28.36v-2.5q0-1.875-2.8500000000000014-3.046666666666667t-5.509999999999998-1.1716666666666669-5.508333333333333 1.1716666666666669-2.8499999999999996 3.046666666666667v2.5h16.71666666666667z m-8.36-17.11q-1.5633333333333326 0-2.6566666666666663 1.0933333333333337t-1.0933333333333337 2.6566666666666663 1.0933333333333337 2.6566666666666663 2.6566666666666663 1.0933333333333337 2.6566666666666663-1.0933333333333337 1.0933333333333337-2.6566666666666663-1.0933333333333337-2.6566666666666663-2.6566666666666663-1.0933333333333337z m13.36-4.61q1.3283333333333331 0 2.3049999999999997 1.0166666666666666t0.9750000000000014 2.341666666666667v20q0 1.326666666666668-0.9766666666666666 2.3416666666666686t-2.306666666666665 1.0166666666666657h-26.713333333333335q-1.330000000000001 0-2.3066666666666675-1.0166666666666657t-0.9766666666666666-2.3400000000000034v-20q0-1.33 0.9766666666666666-2.3450000000000006t2.3050000000000006-1.0166666666666666h26.71666666666667z m-26.72 33.36v-3.3599999999999994h26.71666666666667v3.3599999999999994h-26.715000000000003z m26.72-40v3.3600000000000003h-26.716666666666665v-3.3600000000000003h26.716666666666665z' })
                )
            );
        }
    }]);

    return MdContacts;
}(React.Component);

exports.default = MdContacts;
module.exports = exports['default'];