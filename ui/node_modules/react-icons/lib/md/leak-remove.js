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

var MdLeakRemove = function (_React$Component) {
    _inherits(MdLeakRemove, _React$Component);

    function MdLeakRemove() {
        _classCallCheck(this, MdLeakRemove);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdLeakRemove).apply(this, arguments));
    }

    _createClass(MdLeakRemove, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm25.625 19.296666666666667q4.453333333333337-2.6566666666666663 9.375-2.6566666666666663v3.3599999999999994q-3.671666666666667 0-6.953333333333333 1.7166666666666686z m7.578333333333333 7.578333333333333l-2.6566666666666663-2.6566666666666663q2.1099999999999994-0.8583333333333343 4.453333333333333-0.8583333333333343v3.2833333333333314q-0.9383333333333326 0-1.7966666666666669 0.23333333333333428z m-9.843333333333334-21.875q0 4.921666666666667-2.6566666666666663 9.375l-2.419999999999998-2.421666666666667q1.716666666666665-3.283333333333333 1.716666666666665-6.953333333333333h3.3599999999999994z m-18.36 2.1100000000000003l2.1100000000000003-2.1100000000000003 27.89 27.890000000000008-2.1099999999999923 2.1099999999999923-4.766666666666666-4.766666666666666q-1.4833333333333343 2.1116666666666646-1.4833333333333343 4.766666666666666h-3.2833333333333314q0-3.9066666666666663 2.423333333333332-7.109999999999999l-2.421666666666667-2.3433333333333337q-3.3583333333333414 4.140000000000004-3.3583333333333414 9.453333333333333h-3.3633333333333333q0-6.716666666666669 4.375-11.875l-4.140000000000001-4.140000000000001q-5.154999999999999 4.376666666666669-11.871666666666666 4.376666666666669v-3.361666666666668q5.388333333333334 0 9.530000000000001-3.3599999999999994l-2.4216666666666686-2.423333333333334q-3.2033333333333314 2.424999999999999-7.1083333333333325 2.424999999999999v-3.283333333333333q2.6550000000000002 0 4.763333333333334-1.4833333333333325z m11.64-2.1100000000000003q0 2.3433333333333337-0.8599999999999994 4.453333333333333l-2.656666666666668-2.6566666666666654q0.2333333333333325-0.8600000000000003 0.2333333333333325-1.7966666666666669h3.283333333333335z' })
                )
            );
        }
    }]);

    return MdLeakRemove;
}(React.Component);

exports.default = MdLeakRemove;
module.exports = exports['default'];