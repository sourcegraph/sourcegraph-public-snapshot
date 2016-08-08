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

var MdAcUnit = function (_React$Component) {
    _inherits(MdAcUnit, _React$Component);

    function MdAcUnit() {
        _classCallCheck(this, MdAcUnit);

        return _possibleConstructorReturn(this, Object.getPrototypeOf(MdAcUnit).apply(this, arguments));
    }

    _createClass(MdAcUnit, [{
        key: 'render',
        value: function render() {
            return React.createElement(
                IconBase,
                _extends({ viewBox: '0 0 40 40' }, this.props),
                React.createElement(
                    'g',
                    null,
                    React.createElement('path', { d: 'm36.64000000000001 18.36v3.2833333333333314h-6.953333333333333l5.390000000000001 5.388333333333335-2.3433333333333337 2.4200000000000017-7.733333333333341-7.811666666666667h-3.361666666666668v3.3599999999999994l7.813333333333333 7.733333333333334-2.419999999999998 2.344999999999999-5.390000000000001-5.390000000000001v6.953333333333333h-3.2833333333333314v-6.953333333333333l-5.388333333333334 5.390000000000001-2.42-2.344999999999999 7.809999999999999-7.733333333333334v-3.3599999999999994h-3.3599999999999994l-7.733333333333333 7.813333333333333-2.3449999999999998-2.421666666666667 5.390000000000001-5.391666666666666h-6.953333333333333v-3.2833333333333314h6.953333333333333l-5.390000000000001-5.385000000000002 2.3450000000000006-2.421666666666667 7.7333333333333325 7.809999999999999h3.3599999999999994v-3.3599999999999994l-7.813333333333333-7.7333333333333325 2.42-2.3433333333333337 5.3916666666666675 5.390000000000001v-6.953333333333334h3.2833333333333314v6.953333333333332l5.388333333333335-5.390000000000001 2.4200000000000017 2.3433333333333355-7.810000000000002 7.7333333333333325v3.361666666666668h3.3599999999999994l7.733333333333334-7.813333333333333 2.3433333333333337 2.42-5.390000000000001 5.389999999999999h6.953333333333333z' })
                )
            );
        }
    }]);

    return MdAcUnit;
}(React.Component);

exports.default = MdAcUnit;
module.exports = exports['default'];